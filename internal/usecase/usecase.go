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
	InsertOrder(ctx context.Context, user string, order string) error
	// GetOrders - получение списка загруженных номеров заказов
	GetOrders(ctx context.Context, user string) ([]byte, error)
	// GetBalance - получение текущего баланса пользователя
	GetBalance(ctx context.Context, login string) (float32, float32, error)
	// Debit - запрос на списание средств
	Debit(ctx context.Context, user, numOrder string, sum float32) error
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
func (uc *UseCase) DoInsertOrder(ctx context.Context, user string, order string) error {
	err := uc.r.InsertOrder(ctx, user, order)
	return err
}

func (uc *UseCase) DoGetOrders(ctx context.Context, user string) ([]byte, error) {
	orders, err := uc.r.GetOrders(ctx, user)
	return orders, err
}

func (uc *UseCase) DoGetBalance(ctx context.Context, login string) (float32, float32, error) {
	b, w, err := uc.r.GetBalance(ctx, login)
	return b, w, err
}

func (uc *UseCase) DoDebit(ctx context.Context, user, numOrder string, sum float32) error {
	err := uc.r.Debit(ctx, user, numOrder, sum)
	return err
}
