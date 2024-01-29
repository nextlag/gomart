package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/nextlag/gomart/internal/entity"
	"github.com/nextlag/gomart/pkg/luna"
)

const tick = time.Second * 1

// Withdrawals - cтруктура, предназначенная для возврата клиенту данных о заказах со снятыми бонусами.
type Withdrawals struct {
	Order string    `json:"order"`
	Sum   float32   `json:"sum"`
	Time  time.Time `json:"processed_at"`
}

// Register registers a new user with the provided login and password.
func (uc *UseCase) Register(ctx context.Context, login string, password string) error {
	user := &entity.User{
		Login:    login,
		Password: password,
	}
	db := bun.NewDB(uc.DB, pgdialect.New())

	_, err := db.NewInsert().
		Model(user).
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}

// Auth authenticates a user with the provided login and password.
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

// InsertOrder inserts a new order for the specified user.
func (uc *UseCase) InsertOrder(ctx context.Context, user string, order string) error {
	now := time.Now()

	bonusesWithdrawn := float32(0)

	userOrder := &entity.Orders{
		Users:            user,
		Order:            order,
		Status:           "NEW",
		UploadedAt:       now,
		BonusesWithdrawn: bonusesWithdrawn,
	}
	validOrder := luna.CheckValidOrder(order)
	if !validOrder {
		return uc.Err().ErrOrderFormat
	}

	db := bun.NewDB(uc.DB, pgdialect.New())

	var checkOrder entity.Orders

	err := db.NewSelect().
		Model(&checkOrder).
		Where(`"order" = ?`, order).
		Scan(ctx)
	if errors.Is(err, nil) {
		// Заказ существует
		if checkOrder.Users == user {
			// Заказ принадлежит текущему пользователю
			return uc.Err().ErrThisUser
		}
		// Заказ принадлежит другому пользователю
		return uc.Err().ErrAnotherUser
	}

	// Заказ не существует, вставьте его
	_, err = db.NewInsert().
		Model(userOrder).
		Set("uploaded_at = ?", now.Format(time.RFC3339)).
		Exec(ctx)
	if err != nil {
		return err
	}

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

// GetBalance retrieves the balance and withdrawn amounts for a user.
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
		uc.log.Error("error finding user's balance", err.Error())
		return 0, 0, err
	}

	return user.Balance, user.Withdrawn, nil
}

// Debit выполняет списание бонусов со счета пользователя за определенный заказ.
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
