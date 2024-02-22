// Package usecase provides the application's business logic.
package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/nextlag/gomart/internal/entity"
	"github.com/nextlag/gomart/pkg/luna"
)

const tick = time.Second * 1

// Withdrawals structure designed to return data to the client about orders with removed bonuses.
type Withdrawals struct {
	Order string    `json:"order"`
	Sum   float32   `json:"sum"`
	Time  time.Time `json:"processed_at"`
}

// Register регистрирует нового пользователя с предоставленным логином и паролем.
// Метод начинает транзакцию с базой данных, создает нового пользователя с указанными
// данными и вставляет его в базу данных. После успешной вставки, транзакция фиксируется,
// а если происходит ошибка, транзакция откатывается.
//
// Параметры:
//   - ctx: контекст выполнения запроса.
//   - login: логин нового пользователя.
//   - password: пароль нового пользователя.
//
// Возвращаемое значение:
//   - error: если произошла ошибка во время выполнения запроса или транзакции,
//     возвращается ошибка, в противном случае nil.
func (uc *UseCase) Register(ctx context.Context, login string, password string) error {
	user := &entity.User{
		Login:    login,
		Password: password,
	}
	tx, err := uc.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %v", err)
	}
	defer tx.Rollback() // Откатить транзакцию в случае ошибки
	db := bun.NewDB(uc.DB, pgdialect.New())

	_, err = db.NewInsert().
		Model(user).
		Exec(ctx)

	if err != nil {
		return err
	}
	// Завершить транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}
	return nil
}

// Auth выполняет аутентификацию пользователя по указанному логину и паролю.
// Метод выполняет запрос к базе данных для поиска пользователя с указанными
// учетными данными. Если пользователь существует и его учетные данные совпадают
// с переданными в метод логином и паролем, метод завершается успешно. В противном
// случае возвращается ошибка.
//
// Параметры:
//   - ctx: контекст выполнения запроса.
//   - login: логин пользователя.
//   - password: пароль пользователя.
//
// Возвращаемое значение:
//   - error: в случае успешной аутентификации возвращается nil, в противном случае
//     возвращается ошибка.
func (uc *UseCase) Auth(ctx context.Context, login, password string) error {
	var user entity.User

	db := bun.NewDB(uc.DB, pgdialect.New())

	err := db.NewSelect().
		Model(&user).
		Where("login = ? and password = ?", login, password).
		Scan(ctx)

	if err != nil {
		return err
	}

	return nil
}

// InsertOrder осуществляет вставку нового заказа в базу данных.
// Метод принимает контекст ctx типа context.Context, имя пользователя user и описание заказа order.
// Контекст ctx используется для управления временем жизни операции и для передачи значения времени выполнения, которое должно учитываться при выполнении операции.
// Имя пользователя user представляет собой уникальный идентификатор пользователя, оформляющего заказ.
// Описание заказа order содержит информацию о заказе, которую пользователь хочет добавить в базу данных.
// Возвращает ошибку в случае любого сбоя операции или невозможности выполнить запрос к базе данных.
// Возможные ошибки, которые могут возникнуть включают в себя:
//   - ErrOrderFormat: неверный формат заказа.
//   - ErrThisUser: заказ уже существует и принадлежит текущему пользователю.
//   - ErrAnotherUser: заказ уже существует и принадлежит другому пользователю.
//   - ошибку при начале транзакции.
//   - ошибку при выполнении запроса к базе данных.
//   - ошибку при коммите транзакции.
func (uc *UseCase) InsertOrder(ctx context.Context, user string, order string) error {
	// Получаем текущее время.
	now := time.Now()

	// Инициализируем переменную для отзыва бонусов.
	bonusesWithdrawn := float32(0)

	// Создаем объект заказа.
	userOrder := &entity.Orders{
		Users:            user,
		Order:            order,
		Status:           "NEW",
		UploadedAt:       now,
		BonusesWithdrawn: bonusesWithdrawn,
	}

	// Проверяем валидность заказа.
	validOrder := luna.CheckValidOrder(order)
	if !validOrder {
		return uc.Err().ErrOrderFormat
	}

	// Начинаем транзакцию.
	tx, err := uc.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %v", err)
	}
	defer tx.Rollback()

	// Создаем новый объект для работы с базой данных.
	db := bun.NewDB(uc.DB, pgdialect.New())

	// Проверяем существует ли уже заказ с таким же описанием.
	var checkOrder entity.Orders
	err = db.NewSelect().
		Model(&checkOrder).
		Where(`"order" = ?`, order).
		Scan(ctx)
	if errors.Is(err, nil) {
		// Если заказ существует, проверяем его принадлежность пользователю.
		if checkOrder.Users == user {
			// Заказ принадлежит текущему пользователю.
			return uc.Err().ErrThisUser
		}
		// Заказ принадлежит другому пользователю.
		return uc.Err().ErrAnotherUser
	}

	// Заказ не существует, вставляем его в базу данных.
	_, err = db.NewInsert().
		Model(userOrder).
		Set("uploaded_at = ?", now.Format(time.RFC3339)).
		Exec(ctx)
	if err != nil {
		return err
	}

	// Коммит транзакции.
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	// Возвращаем nil в случае успешного выполнения операции.
	return nil
}

// GetOrders retrieves all orders for a specific user.
func (uc *UseCase) GetOrders(ctx context.Context, user string) ([]byte, error) {
	var (
		allOrders []entity.Orders
		userOrder entity.Orders
	)
	db := bun.NewDB(uc.DB, pgdialect.New())

	rows, err := db.NewSelect().
		Model(&userOrder).
		Where("users = ?", user).
		Order("uploaded_at ASC").
		Rows(ctx)
	if err != nil {
		return nil, err
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var en entity.Orders
		err = rows.Scan(&en.Users, &en.Order, &en.Status, &en.Accrual, &en.UploadedAt, &en.BonusesWithdrawn)
		if err != nil {
			return nil, err
		}

		allOrders = append(allOrders, entity.Orders{
			Order:      en.Order,
			Status:     en.Status,
			Accrual:    en.Accrual,
			UploadedAt: en.UploadedAt,
		})
	}

	result, err := json.Marshal(allOrders)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetBalance возвращает текущий баланс и сумму снятых средств пользователя по указанному логину.
// Метод принимает контекст ctx типа context.Context и логин пользователя login.
// Контекст ctx используется для управления временем жизни операции и для передачи значения времени выполнения, которое должно учитываться при выполнении операции.
// Логин пользователя login является уникальным идентификатором пользователя, для которого нужно получить баланс.
// Возвращает текущий баланс и сумму снятых средств пользователя. В случае успешного выполнения запроса, метод возвращает два значения типа float32:
//   - текущий баланс
//   - сумму снятых средств.
//     Если произошла ошибка при выполнении запроса к базе данных, метод возвращает ошибку.
//     В случае ошибки, текущий баланс и сумму снятых средств считаются нулевыми.
func (uc *UseCase) GetBalance(ctx context.Context, login string) (float32, float32, error) {
	var user entity.User

	db := bun.NewDB(uc.DB, pgdialect.New())

	// Выполнение SELECT запроса к базе данных для получения бонусов по указанному логину
	err := db.NewSelect().
		Model(&user).
		ColumnExpr("balance, withdrawn").
		Where("login = ?", login).
		Scan(ctx)
	if err != nil {
		fmt.Printf("error finding user's balance: %v", err)
		return 0, 0, err
	}

	return user.Balance, user.Withdrawn, nil
}

// Debit debits bonuses from the user's account for a specific order.
func (uc *UseCase) Debit(ctx context.Context, user string, order string, sum float32) error {
	// Проверка корректности номера заказа
	validOrder := luna.CheckValidOrder(order)
	if !validOrder {
		// Если заказ некорректен, возвращает ErrOrderFormat.
		return uc.Err().ErrOrderFormat
	}

	// Получение текущего баланса пользователя
	balance, _, err := uc.GetBalance(ctx, user)
	if err != nil {
		return err
	}
	if balance < sum {
		// Если на счету пользователя недостаточно средств, возвращает ошибку
		return uc.Err().ErrNoBalance
	}

	// Инициализация подключения к базе данных
	db := bun.NewDB(uc.DB, pgdialect.New())

	// Проверка существования заказа в базе данных
	checkOrder := entity.Orders{}
	now := time.Now()

	err = db.NewSelect().
		Model(checkOrder).
		Where(`"order" = ?`, order).
		Scan(ctx)
	if !errors.Is(err, nil) {
		// Заказ не существует, добавляем новый заказ в базу данных
		_, err := db.NewInsert().
			Model(&entity.Orders{
				Users:            user,
				Order:            order,
				UploadedAt:       now,
				Status:           "NEW",
				BonusesWithdrawn: sum,
			}).
			Set("uploaded_at = ?", now.Format(time.RFC3339)).
			Exec(ctx)
		if !errors.Is(err, nil) {
			return err
		}
	}

	// Проверка принадлежности заказа текущему пользователю или другому
	if checkOrder.Users != user && checkOrder.Order == order {
		// Если заказ существует и принадлежит другому пользователю, возвращает ErrAnotherUser.
		return uc.Err().ErrAnotherUser
	} else if checkOrder.Users == user && checkOrder.Order == order {
		// Если заказ существует и принадлежит текущему пользователю, возвращает ErrThisUser.
		return uc.Err().ErrThisUser
	}

	// Если заказ существует, обновляет баланс пользователя и добавляет запись о списании в базу данных.
	_, err = db.NewUpdate().
		TableExpr("users").
		Set("balance = ?", balance-sum).
		Set("withdrawn = withdrawn + ?", sum).
		Where("login = ?", user).
		Exec(ctx)
	if !errors.Is(err, nil) {
		return err
	}

	return nil
}

// GetWithdrawals method for getting all orders with bonuses removed for a specific user.
func (uc *UseCase) GetWithdrawals(ctx context.Context, user string) ([]byte, error) {
	var (
		allOrders []Withdrawals
		userOrder entity.Orders
	)

	// Инициализация подключения к базе данных
	db := bun.NewDB(uc.DB, pgdialect.New())

	// Выборка всех заказов пользователя, где bonuses_withdrawn != 0
	rows, err := db.NewSelect().
		Model(&userOrder).
		Where("users = ? and bonuses_withdrawn != 0", user).
		Order("uploaded_at ASC").
		Rows(ctx)
	rows.Err()

	if err != nil {
		return nil, err
	}

	noRows := true
	for rows.Next() {
		noRows = false
		var orderRow entity.Orders
		if err = rows.Scan(
			&orderRow.Users,
			&orderRow.Order,
			&orderRow.Status,
			&orderRow.Accrual,
			&orderRow.UploadedAt,
			&orderRow.BonusesWithdrawn,
		); err != nil {
			return nil, err
		}

		allOrders = append(allOrders, Withdrawals{
			Order: orderRow.Order,
			Sum:   orderRow.BonusesWithdrawn,
			Time:  orderRow.UploadedAt,
		})
	}

	if noRows {
		return nil, ErrNoRows
	}

	result, err := json.Marshal(allOrders)
	if err != nil {
		return nil, err
	}
	return result, nil
}
