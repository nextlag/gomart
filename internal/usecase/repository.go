package usecase

import (
	"context"
	"log/slog"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/nextlag/gomart/internal/entity"
	"github.com/nextlag/gomart/internal/repository/psql"
)

type Storage struct {
	*psql.Postgres
	*slog.Logger
}

func NewStorage(db *psql.Postgres, log *slog.Logger) *Storage {
	return &Storage{db, log}
}

func (s *Storage) Register(ctx context.Context, login string, password string) error {
	credentials := &entity.User{
		Login:    login,
		Password: password,
	}
	db := bun.NewDB(s.Postgres.DB, pgdialect.New())

	_, err := db.NewInsert().
		Model(credentials).
		Exec(ctx)

	if err != nil {
		s.Error("Error writing data: ", err.Error())
		return err
	}

	return nil
}
