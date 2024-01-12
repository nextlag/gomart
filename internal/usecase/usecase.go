package usecase

import (
	"context"
	"net/http"

	"github.com/nextlag/gomart/internal/entity"
)

//go:generate mockgen -destination=mocks/mocks.go -package=mocks github.com/nextlag/gomart/internal/usecase Repository
type Repository interface {
	// Register - регистрация пользователя
	Register(ctx context.Context, login, password string) error
	// Auth — аутентификация пользователя.
	Auth(ctx context.Context, login, password string) error
	// InsertOrder - загрузка номера заказа.
	InsertOrder(ctx context.Context, login string, order string) error
	// GetOrders - получение списка загруженных номеров заказов
	GetOrders(ctx context.Context, login string) ([]byte, error)
}

type UseCase struct {
	r Repository // interface Repository
	e *entity.AllEntity
}

func New(r Repository) *UseCase {
	e := &entity.AllEntity{}
	return &UseCase{r, e}
}

func (uc *UseCase) GetEntity() *entity.AllEntity {
	return uc.e
}

func (uc *UseCase) DoRegister(ctx context.Context, login, password string, _ *http.Request) error {
	err := uc.r.Register(ctx, login, password)
	return err
}
func (uc *UseCase) DoAuth(ctx context.Context, login, password string, _ *http.Request) error {
	err := uc.r.Auth(ctx, login, password)
	return err
}
func (uc *UseCase) DoInsertOrder(ctx context.Context, login string, order string) error {
	err := uc.r.InsertOrder(ctx, login, order)
	return err
}

func (uc *UseCase) DoGetOrders(ctx context.Context, login string) ([]byte, error) {
	orders, err := uc.r.GetOrders(ctx, login)
	return orders, err
}
