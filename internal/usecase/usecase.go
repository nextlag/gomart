package usecase

import (
	"context"
	"net/http"

	"github.com/nextlag/gomart/internal/entity"
)

type Repository interface {
	// Register - регистрация нового пользователя
	Register(ctx context.Context, login, password string) error
	// Auth — проверяет, есть ли совпадение в базе по логину и паролю.
	Auth(ctx context.Context, login, password string) error
	// InsertOrder - используется для вставки информации о заказе в базу данных.
	InsertOrder(ctx context.Context, login string, order string) error
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
	err := uc.r.Register(ctx, login, password)
	return err
}
func (uc *UseCase) DoAuth(ctx context.Context, login, password string, r *http.Request) error {
	err := uc.r.Auth(ctx, login, password)
	return err
}
func (uc *UseCase) DoInsertOrder(ctx context.Context, login string, order string) error {
	err := uc.r.InsertOrder(ctx, login, order)
	return err
}
