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

// ErrAuth authenticates a user with the provided login and password.
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

// Debit - обновляя баланс и снимаемые суммы.
func (uc *UseCase) Debit(ctx context.Context, user, order string, sum float32) error {
	validOrder := luna.CheckValidOrder(order)
	if !validOrder {
		return uc.Err().ErrOrderFormat
	}

	// Получение текущего баланса пользователя
	balance, _, err := uc.GetBalance(ctx, user)
	if err != nil {
		return err
	}

	// Проверка наличия достаточного баланса для списания бонусов
	if balance < sum {
		return uc.Err().ErrNoBalance
	}

	// Инициализация подключения к базе данных
	db := bun.NewDB(uc.DB, pgdialect.New())

	// Получение текущей даты и времени
	now := time.Now()

	// Создание объекта заказа пользователя
	userOrder := &entity.Orders{
		Users:            user,
		Order:            order,
		Status:           "NEW",
		UploadedAt:       now,
		BonusesWithdrawn: sum,
	}

	// Проверка существования заказа в базе данных
	err = db.NewSelect().
		Model(&userOrder).
		Where(`"order" = ?`, order).
		Scan(ctx)
	if err != nil {
		// Заказ не существует, добавляем новый заказ в базу данных
		_, err = db.NewInsert().
			Model(userOrder).
			Set("uploaded_at = ?", now.Format(time.RFC3339)).
			Exec(ctx)
		// if err != nil {
		// 	// Заказ существует
		// 	if userOrder.Users == user {
		// 		// Заказ принадлежит текущему пользователю
		// 		return uc.Err().ErrThisUser
		// 	}
		// 	// Заказ принадлежит другому пользователю
		// 	return uc.Err().ErrAnotherUser
		// }
		return err
	}
	var checkOrder entity.Orders
	if checkOrder.Users != user && checkOrder.Order == order {
		return uc.Err().ErrAnotherUser
	} else if checkOrder.Users == user && checkOrder.Order == order {
		return uc.Err().ErrThisUser
	}

	// Обновление баланса пользователя после списания бонусов
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
		allOrders []entity.Withdrawals
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

		allOrders = append(allOrders, entity.Withdrawals{
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
