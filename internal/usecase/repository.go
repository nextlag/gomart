package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/nextlag/gomart/internal/entity"
	"github.com/nextlag/gomart/internal/repository/psql"
	"github.com/nextlag/gomart/pkg/luna"
)

type Storage struct {
	*AllErr
	*psql.Postgres
	*slog.Logger
}

func NewStorage(er *AllErr, db *psql.Postgres, log *slog.Logger) *Storage {
	return &Storage{er, db, log}
}

func (s *Storage) Register(ctx context.Context, login string, password string) error {
	user := &entity.User{
		Login:    login,
		Password: password,
	}
	db := bun.NewDB(s.Postgres.DB, pgdialect.New())

	_, err := db.NewInsert().
		Model(user).
		Exec(ctx)

	if err != nil {
		s.Error("error writing data: ", "error usecase Register", err.Error())
		return err
	}

	return nil
}

func (s *Storage) Auth(ctx context.Context, login, password string) error {
	var user entity.User

	db := bun.NewDB(s.Postgres.DB, pgdialect.New())

	err := db.NewSelect().
		Model(&user).
		Where("login = ? and password = ?", login, password).
		Scan(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) InsertOrder(ctx context.Context, login string, order string) error {
	now := time.Now()

	bonusesWithdrawn := float32(0)

	userOrder := &entity.Order{
		Login:      login,
		Order:      order,
		UploadedAt: now.Format(time.RFC3339),
		Status:     "NEW",
		Bonuses:    &bonusesWithdrawn,
	}
	validOrder := luna.CheckValidOrder(order)
	if !validOrder {
		s.Logger.Debug("InsertOrder", "no valid", validOrder, "status", "invalid order format")
		return s.OrderFormat
	}

	db := bun.NewDB(s.DB, pgdialect.New())

	var checkOrder entity.Order

	err := db.NewSelect().
		Model(&checkOrder).
		Where(`"order" = ?`, order).
		Scan(ctx)
	if err == nil {
		// Заказ существует
		if checkOrder.Login == login {
			// Заказ принадлежит текущему пользователю
			return s.ThisUser
		}
		// Заказ принадлежит другому пользователю
		return s.AnotherUser
	}

	// Заказ не существует, вставьте его
	_, err = db.NewInsert().
		Model(userOrder).
		Exec(ctx)
	if err != nil {
		s.Error("error writing data: ", "error usecase InsertOrder", err.Error())
		return err
	}

	return nil
}

func (s *Storage) GetOrders(ctx context.Context, login string) ([]entity.Order, error) {
	var allOrders []entity.Order
	order := entity.Order{}

	db := bun.NewDB(s.Postgres.DB, pgdialect.New())

	rows, err := db.NewSelect().
		Model(&order).
		Where("login = ?", login).
		Order("uploaded_at ASC").
		Rows(ctx)
	if err != nil {
		s.Logger.Error("error getting data", "GetOrders", err.Error())
		return nil, err
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var en entity.Order
		err = rows.Scan(
			&en.Login,
			&en.Order,
			&en.Status,
			&en.UploadedAt,
			&en.Bonuses,
			&en.Accrual,
		)
		if err != nil {
			s.Logger.Error("error scanning data", "GetOrders", err.Error())
			return nil, err
		}

		// Логируем данные, чтобы проверить, что они получены корректно
		s.Logger.Debug("Got order", "Order", en.Order, "Status", en.Status, "UploadedAt", en.UploadedAt)

		allOrders = append(allOrders, entity.Order{
			Order:      en.Order,
			UploadedAt: en.UploadedAt,
			Status:     en.Status,
			Accrual:    en.Accrual,
		})
	}
	s.Logger.Debug("allOrders", "Orders", allOrders)
	return allOrders, nil
}
