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
		uc.log.Error("error writing data: ", "usecase Register", err.Error())
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
		Number:           order,
		UploadedAt:       now.Format(time.RFC3339),
		Status:           "NEW",
		BonusesWithdrawn: bonusesWithdrawn,
	}
	validOrder := luna.CheckValidOrder(order)
	if !validOrder {
		uc.log.Debug("InsertOrder", "no valid", validOrder, "status", "invalid order format")
		return uc.er.OrderFormat
	}

	db := bun.NewDB(uc.DB, pgdialect.New())

	var checkOrder entity.Orders

	err := db.NewSelect().
		Model(&checkOrder).
		Where(`"number" = ?`, order).
		Scan(ctx)
	if errors.Is(err, nil) {
		// Заказ существует
		if checkOrder.Users == user {
			// Заказ принадлежит текущему пользователю
			uc.log.Debug("current user order", "this user", checkOrder.Users)
			return uc.er.ThisUser
		}
		// Заказ принадлежит другому пользователю
		uc.log.Debug("another user order", "another user", checkOrder.Users)
		return uc.er.AnotherUser
	}

	// Заказ не существует, вставьте его
	_, err = db.NewInsert().
		Model(userOrder).
		Exec(ctx)
	if err != nil {
		uc.log.Error("error writing data", "usecase InsertOrder", err.Error())
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
		uc.log.Error("error getting data", "usecase GetOrders", err.Error())
		return nil, err
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var en entity.Orders
		err = rows.Scan(&en.Users, &en.Number, &en.Status, &en.Accrual, &en.UploadedAt, &en.BonusesWithdrawn)
		if err != nil {
			uc.log.Error("error scanning data", "usecase GetOrders", err.Error())
			return nil, err
		}

		allOrders = append(allOrders, entity.Orders{
			Number:     en.Number,
			Status:     en.Status,
			Accrual:    en.Accrual,
			UploadedAt: en.UploadedAt,
		})
	}

	result, err := json.Marshal(allOrders)
	if err != nil {
		uc.log.Error("error marshaling allOrders", "usecase GetOrders", err.Error())
		return nil, err
	}
	return result, nil
}

// GetBalance retrieves the balance and withdrawn amounts for a user.
func (uc *UseCase) GetBalance(ctx context.Context, login string) (float32, float32, error) {
	// Инициализация переменной для хранения баланса
	var balance, withdrawn float32
	// Создание экземпляра объекта для взаимодействия с базой данных
	db := bun.NewDB(uc.DB, pgdialect.New())

	// Выполнение SELECT запроса к базе данных для получения бонусов по указанному логину
	err := db.NewSelect().
		TableExpr("users").
		ColumnExpr("balance, withdrawn").
		Where("login = ?", login).
		Scan(ctx, &balance, &withdrawn)
	if err != nil {
		// Возвращаем другие ошибки
		uc.log.Error("error while scanning data", "usecase GetBalance", err.Error())
		return 0, 0, err
	}

	return balance, withdrawn, nil
}

// Debit processes the debit operation for a user, updating the balance and withdrawn amounts.
func (uc *UseCase) Debit(ctx context.Context, user, order string, sum float32) error {
	// Получение текущего баланса пользователя
	var checkOrder entity.Orders
	balance, _, err := uc.GetBalance(ctx, user)
	if err != nil {
		uc.log.Error("error get balance from GetBalance method", "usecase Debit", err.Error())
		return err
	}
	// Проверка наличия достаточного баланса для списания бонусов
	if balance < sum {
		return uc.er.NoBalance
	}

	// Инициализация подключения к базе данных
	db := bun.NewDB(uc.DB, pgdialect.New())

	// Получение текущей даты и времени
	now := time.Now()

	// Создание объекта заказа пользователя
	userOrder := &entity.Orders{
		Users:            user,
		Number:           order,
		Status:           "NEW",
		UploadedAt:       now.Format(time.RFC3339),
		BonusesWithdrawn: sum,
	}

	// Проверка существования заказа в базе данных
	err = db.NewSelect().
		Model(&userOrder).
		Where(`"number" = ?`, order).
		Scan(ctx)
	if errors.Is(err, nil) {
		// Заказ существует
		if checkOrder.Users == user {
			// Заказ принадлежит текущему пользователю
			return uc.er.ThisUser
		}
		// Заказ принадлежит другому пользователю
		return uc.er.AnotherUser
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
