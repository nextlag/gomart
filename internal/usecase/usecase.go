package usecase

import (
	"context"
	"fmt"

	"github.com/nextlag/gomart/internal/entity"
)

type Repository interface {
	GetBalance(ctx context.Context) error
}

type UseCase struct {
	repo Repository // interface Repository
	e    *entity.Entity
}

func New(r Repository) *UseCase {
	e := &entity.Entity{}
	return &UseCase{r, e}
}

func NewEntity(uc UseCase) *entity.Entity {
	return uc.e
}

func (uc *UseCase) Do(ctx context.Context) error {
	err := uc.repo.GetBalance(ctx)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}
	return nil
}
