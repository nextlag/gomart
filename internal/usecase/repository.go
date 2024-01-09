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
	*psql.Postgres
	*slog.Logger
}

func NewStorage(db *psql.Postgres, log *slog.Logger) *Storage {
	return &Storage{db, log}
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
		Login:            login,
		Order:            order,
		UploadedAt:       now.Format(time.RFC3339),
		Status:           "NEW",
		BonusesWithdrawn: &bonusesWithdrawn,
	}
	validOrder := luna.CheckValidOrder(order)
	if !validOrder {
		return Status().OrderFormat
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
			return Status().ThisUser
		}
		// Заказ принадлежит другому пользователю
		return Status().AnotherUser
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
