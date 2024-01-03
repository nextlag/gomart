package usecase

import (
	"context"
	"fmt"

	"github.com/nextlag/gomart/internal/entity"
)

type Repository interface {
	Register(ctx context.Context, login string, password string) error
}

type UseCase struct {
	r Repository // interface Repository
	e *entity.Entity
}

func New(r Repository) *UseCase {
	e := &entity.Entity{}
	return &UseCase{r, e}
}

func NewEntity(uc UseCase) *entity.Entity {
	return uc.e
}

func (uc *UseCase) DoRegister(ctx context.Context, login, password string) error {
	err := uc.r.Register(ctx, login, password)
	if err != nil {
		return fmt.Errorf("failed to push registration data %w", err.Error())
	}
	return nil
}
