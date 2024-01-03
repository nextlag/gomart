package usecase

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nextlag/gomart/internal/entity"
)

type Repository interface {
	// Register - регистрация нового пользователя
	Register(ctx context.Context, login, password string) error
	// Auth — проверяет, есть ли совпадение в базе по логину и паролю.
	Auth(ctx context.Context, login, password string) error
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

func (uc *UseCase) DoRegister(ctx context.Context, login, password string, r *http.Request) error {
	switch {
	case r.URL.Path == "/api/user/register":
		err := uc.r.Register(ctx, login, password)
		if err != nil {
			return fmt.Errorf("failed to push registration data %v", err.Error())
		}
	case r.URL.Path == "/api/user/login":
		err := uc.r.Auth(ctx, login, password)
		if err != nil {
			return fmt.Errorf("failed to push registration data %v", err.Error())
		}

	}
	return nil
}
