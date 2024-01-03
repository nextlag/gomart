package usecase

import (
	"context"

	"github.com/nextlag/gomart/internal/repository/psql"
)

type Storage struct {
	*psql.Postgres
}

func NewStorage(db *psql.Postgres) *Storage {
	return &Storage{db}
}

func (s *Storage) Register(ctx context.Context, login string, password string) error {
	return nil
}
