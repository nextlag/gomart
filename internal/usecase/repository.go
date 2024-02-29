// Package usecase provides the application's business logic.
package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/nextlag/gomart/internal/entity"
	"github.com/nextlag/gomart/pkg/logger/l"
	"github.com/nextlag/gomart/pkg/luna"
)

const tick = time.Second * 1

const (
	insertUser = `
		INSERT INTO "users" (login, password, balance, withdrawn)
		VALUES ($1, $2, 0, 0) RETURNING login, password, balance, withdrawn
	`
	updateUser = `
		UPDATE users
		SET balance = balance - $1, withdrawn = withdrawn + $1
		WHERE login = $2
	`
	selectUser = `
		SELECT login, password, balance, withdrawn
		FROM users
		WHERE login = $1 AND password = $2
	`
	selectOrder = `
		SELECT user_name, "order", status, accrual, uploaded_at, bonuses_withdrawn
		FROM orders
		WHERE "order" = $1
	`
	insertOrder = `
		INSERT INTO orders (user_name, "order", status, accrual, uploaded_at, bonuses_withdrawn) 
		VALUES ($1, $2, 'NEW', 0, $3, 0) 
		RETURNING user_name, "order", status, accrual, uploaded_at, bonuses_withdrawn
	`
	insertOrderWithdrawn = `
		INSERT INTO "orders" (user_name, "order", status, accrual, uploaded_at, bonuses_withdrawn)
		VALUES ($1, $2, 'NEW', 0, $3, $4)
	`
	selectOrders = `
		SELECT "order", status, accrual, uploaded_at
		FROM orders
		WHERE user_name = $1
		ORDER BY uploaded_at ASC
	`

	selectOrderWithdrawals = `
		SELECT "order", bonuses_withdrawn, uploaded_at
		FROM orders
		WHERE user_name = $1 AND bonuses_withdrawn != 0
		ORDER BY uploaded_at ASC
	`
)

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
//
//   - error: если произошла ошибка во время выполнения запроса или транзакции,
//     возвращается ошибка, в противном случае nil.
func (uc *UseCase) Register(ctx context.Context, login, password string) error {
	var eUsers entity.User

	tx, err := uc.DB.BeginTx(ctx, nil)
	if err != nil {
		l.L(ctx).Error("begin transaction", l.ErrAttr(err))
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		err = tx.Commit()
		if err != nil {
			err = fmt.Errorf("commit transaction: %v", err)
		}
	}()

	err = uc.DB.QueryRowContext(ctx, insertUser, login, password).Scan(
		&eUsers.Login,
		&eUsers.Password,
		&eUsers.Balance,
		&eUsers.Withdrawn,
	)

	if err != nil {
		l.L(ctx).Error("error pushing data in table users", l.ErrAttr(err))
		return err
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

	err := uc.DB.QueryRowContext(ctx, selectUser, login, password).Scan(
		&user.Login,
		&user.Password,
		&user.Balance,
		&user.Withdrawn,
	)
	if err != nil {
		return err // Пользователь с таким логином и паролем не найден
	}

	return nil // Пользователь успешно аутентифицирован
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
func (uc *UseCase) InsertOrder(ctx context.Context, user, order string) error {
	now := time.Now()
	userOrder := &entity.Order{
		UserName:   user,
		Order:      order,
		UploadedAt: now,
	}
	validOrder := luna.CheckValidOrder(order)
	if !validOrder {
		return uc.Err().ErrOrderFormat
	}

	// Начинаем транзакцию.
	tx, err := uc.DB.BeginTx(ctx, nil)
	if err != nil {
		l.L(ctx).Error("begin transaction", l.ErrAttr(err))
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		err = tx.Commit()
		if err != nil {
			err = fmt.Errorf("commit transaction: %v", err)
		}
	}()

	// Проверяем существует ли уже заказ с таким же описанием.
	var existingOrder entity.Order
	err = uc.DB.QueryRowContext(ctx, selectOrder, order).Scan(
		&existingOrder.UserName,
		&existingOrder.Order,
		&existingOrder.Status,
		&existingOrder.Accrual,
		&existingOrder.UploadedAt,
		&existingOrder.BonusesWithdrawn,
	)
	if err == nil {
		// Если заказ существует, проверяем его принадлежность пользователю.
		if existingOrder.UserName == user {
			// Заказ принадлежит текущему пользователю.
			return uc.Err().ErrThisUser
		}
		// Заказ принадлежит другому пользователю.
		return uc.Err().ErrAnotherUser
	} else if err != sql.ErrNoRows {
		return err
	}

	// Заказ не существует, вставляем его в базу данных.
	err = uc.DB.QueryRowContext(ctx, insertOrder, user, order, now).Scan(
		&userOrder.UserName,
		&userOrder.Order,
		&userOrder.Status,
		&userOrder.Accrual,
		&userOrder.UploadedAt,
		&userOrder.BonusesWithdrawn,
	)
	if err != nil {
		return err
	}
	return nil
}

// GetOrders возвращает заказы пользователя в формате JSON.
// Метод принимает контекст ctx типа context.Context и имя пользователя user.
// Контекст ctx используется для управления временем жизни операции и для передачи значения времени выполнения, которое должно учитываться при выполнении операции.
// Имя пользователя user является уникальным идентификатором пользователя, для которого нужно получить заказы.
// Возвращает список заказов пользователя в формате JSON. Каждый заказ представлен объектом с полями:
//   - Order (описание заказа)
//   - Status (статус заказа)
//   - Accrual (начисление)
//   - UploadedAt (дата загрузки заказа)
//
// Если произошла ошибка при выполнении запроса к базе данных или при преобразовании результатов в JSON, метод возвращает ошибку.
func (uc *UseCase) GetOrders(ctx context.Context, user string) ([]byte, error) {
	var allOrders []entity.Order

	// Выполняем запрос к базе данных.
	rows, err := uc.DB.QueryContext(ctx, selectOrders, user)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %v", err)
	}
	defer rows.Close()

	// Обрабатываем результаты запроса.
	for rows.Next() {
		var order entity.Order
		if err = rows.Scan(&order.Order, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, err
		}
		allOrders = append(allOrders, order)
	}

	// Проверяем наличие ошибок после завершения перебора результатов.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Преобразуем полученные заказы в JSON.
	result, err := json.Marshal(allOrders)
	if err != nil {
		return nil, err
	}

	// Возвращаем JSON-представление заказов.
	log.Print(string(result))
	return result, nil
}

// GetBalance возвращает текущий баланс и сумму снятых средств пользователя по указанному логину.
// Метод принимает контекст ctx типа context.Context и логин пользователя login.
// Контекст ctx используется для управления временем жизни операции и для передачи значения времени выполнения, которое должно учитываться при выполнении операции.
// Логин пользователя login является уникальным идентификатором пользователя, для которого нужно получить баланс.
// Возвращает текущий баланс и сумму снятых средств пользователя. В случае успешного выполнения запроса, метод возвращает два значения типа float32:
//   - текущий баланс
//   - сумму снятых средств.
//
// Если произошла ошибка при выполнении запроса к базе данных, метод возвращает ошибку.
// В случае ошибки, текущий баланс и сумму снятых средств считаются нулевыми.
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

// Debit осуществляет списание средств с баланса пользователя и добавление информации о заказе в базу данных.
// Метод принимает контекст ctx типа context.Context, имя пользователя user, номер заказа order и сумму списания sum.
// Контекст ctx используется для управления временем жизни операции и для передачи значения времени выполнения, которое должно учитываться при выполнении операции.
// Имя пользователя user является уникальным идентификатором пользователя, чей баланс будет уменьшен на сумму списания.
// Номер заказа order представляет описание заказа, для которого будет произведено списание.
// Сумма списания sum указывает количество средств, которое будет списано с баланса пользователя.
// Возвращает ошибку в случае любого сбоя операции.
//
// Возможные ошибки:
//   - ErrOrderFormat: некорректный формат номера заказа.
//   - ErrNoBalance: недостаточно средств на балансе пользователя.
//   - ErrBeginningTransaction: ошибка при начале транзакции.
//   - ErrExecutingDatabaseQuery: ошибка при выполнении запроса к базе данных.
//   - ErrCommittingTransaction: ошибка при коммите транзакции.
//   - ErrAnotherUser: заказ существует и принадлежит другому пользователю.
//   - ErrThisUser: заказ существует и принадлежит текущему пользователю.
func (uc *UseCase) Debit(ctx context.Context, user, order string, sum float32) error {
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

	// Если на счету пользователя недостаточно средств, возвращает ошибку
	if balance < sum {
		return uc.Err().ErrNoBalance
	}

	tx, err := uc.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction Debit method: %v", err)
	}
	defer tx.Rollback()

	// Проверка существования заказа в базе данных
	var existingOrder entity.Order
	err = uc.DB.QueryRowContext(ctx, selectOrder, order).Scan(&existingOrder.UserName, &existingOrder.Order)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("error checking order existence: %v", err)
	}

	if existingOrder.UserName != user && existingOrder.Order == order {
		// Если заказ существует и принадлежит другому пользователю, возвращает ErrAnotherUser.
		return uc.Err().ErrAnotherUser
	} else if existingOrder.UserName == user && existingOrder.Order == order {
		// Если заказ существует и принадлежит текущему пользователю, возвращает ErrThisUser.
		return uc.Err().ErrThisUser
	}

	// Обновляем баланс пользователя и добавляем запись о списании в базу данных.
	now := time.Now()
	_, err = uc.DB.ExecContext(ctx, updateUser, sum, user)
	if err != nil {
		return fmt.Errorf("error updating user balance: %v", err)
	}

	_, err = uc.DB.ExecContext(ctx, insertOrderWithdrawn, user, order, now, sum)

	if err != nil {
		return fmt.Errorf("error inserting order: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction Debit method: %v", err)
	}
	return nil
}

// Withdrawals structure designed to return data to the client about orders with removed bonuses.
type Withdrawals struct {
	Order string    `json:"order"`
	Sum   float32   `json:"sum"`
	Time  time.Time `json:"processed_at"`
}

// GetWithdrawals возвращает список снятий бонусов пользователя в формате JSON.
// Метод принимает контекст ctx типа context.Context и имя пользователя user.
// Контекст ctx используется для управления временем жизни операции и для передачи значения времени выполнения, которое должно учитываться при выполнении операции.
// Имя пользователя user является уникальным идентификатором пользователя, для которого нужно получить список снятий бонусов.
// Возвращает список снятий бонусов пользователя в формате JSON. Каждое снятие бонусов представлено объектом с полями Order (описание заказа), Sum (сумма снятия), Time (время снятия).
// Если не найдено ни одного снятия бонусов для указанного пользователя, метод возвращает ошибку ErrNoRows.
// В случае успешного выполнения операции, метод возвращает список снятий бонусов пользователя в формате JSON и nil.
// В случае любой другой ошибки, возникшей при выполнении запроса к базе данных или при преобразовании результатов в JSON, метод возвращает ошибку.
func (uc *UseCase) GetWithdrawals(ctx context.Context, user string) ([]byte, error) {
	var allWithdrawals []Withdrawals

	rows, err := uc.DB.QueryContext(ctx, selectOrderWithdrawals, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	noRows := true
	for rows.Next() {
		noRows = false
		var order string
		var sum float32
		var time time.Time
		if err := rows.Scan(&order, &sum, &time); err != nil {
			return nil, err
		}

		allWithdrawals = append(allWithdrawals, Withdrawals{
			Order: order,
			Sum:   sum,
			Time:  time,
		})
	}

	// Проверка на наличие ошибок после завершения итерации по строкам
	if err = rows.Err(); err != nil {
		return nil, err
	}

	if noRows {
		return nil, ErrNoRows
	}

	result, err := json.Marshal(allWithdrawals)
	if err != nil {
		return nil, err
	}
	return result, nil
}
