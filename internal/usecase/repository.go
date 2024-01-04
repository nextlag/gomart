package usecase

import (
	"context"
	"errors"
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

var (
	// ErrAnotherUser - заказ загружен другим пользователем
	ErrAnotherUser = errors.New("the order number has already been uploaded by another user")
	// ErrThisUser - дубль заказа текущего пользователем
	ErrThisUser = errors.New("the order number has already been uploaded by this user")
	// ErrOrderFormat - ошибка формата заказа
	ErrOrderFormat = errors.New("invalid order format")
	// ErrNoBalance - недостаточно баланса.
	ErrNoBalance = errors.New("not enough balance")
	// ErrNoRows - строки не найдены
	ErrNoRows = errors.New("no rows were found")
)

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
		s.Error("error writing data: ", "error Register", err.Error())
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
		return ErrOrderFormat
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
			return ErrThisUser
		}
		// Заказ принадлежит другому пользователю
		return ErrAnotherUser
	}

	// Заказ не существует, вставьте его
	_, err = db.NewInsert().
		Model(userOrder).
		Exec(ctx)
	if err != nil {
		s.Error("Error writing data: ", "error InsertOrder", err.Error())
		return err
	}

	return nil
}

// func (s *Storage) InsertOrder(ctx context.Context, login string, order string) error {
// 	now := time.Now()
//
// 	bonusesWithdrawn := float32(0)
//
// 	userOrder := &entity.Order{
// 		Login:            login,
// 		Order:            order,
// 		UploadedAt:       now.Format(time.RFC3339),
// 		Status:           "NEW",
// 		BonusesWithdrawn: &bonusesWithdrawn,
// 	}
//
// 	db := bun.NewDB(s.DB, pgdialect.New())
//
// 	var checkOrder entity.Order
//
// 	err := db.NewSelect().
// 		Model(&checkOrder).
// 		Where(`"order" = ?`, order).
// 		Scan(ctx)
// 	if err != nil {
// 		_, err := db.NewInsert().
// 			Model(userOrder).
// 			Exec(ctx)
// 		if err != nil {
// 			s.Error("error writing data: ", "error InsertOrder", err.Error())
// 			return err
// 		}
//
// 	}
// 	if checkOrder.Login != login && checkOrder.Order == order {
// 		return ErrAlreadyLoadedOrder
// 	} else if checkOrder.Login == login && checkOrder.Order == order {
// 		return ErrThisUser
// 	}
//
// 	return nil
// }
