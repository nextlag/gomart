package usecase

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/nextlag/gomart/internal/entity"
)

type Logger interface {
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
	Error(msg string, args ...any)
}

//go:generate mockgen -destination=mocks/mocks.go -package=mocks github.com/nextlag/gomart/internal/usecase Repository
type Repository interface {
	// Register - регистрация пользователя
	Register(ctx context.Context, login, password string) error
	// Auth ErrAuth — аутентификация пользователя.
	Auth(ctx context.Context, login, password string) error
	// InsertOrder - загрузка номера заказа.
	InsertOrder(ctx context.Context, user string, order string) error
	// GetOrders ErrGetOrders - получение списка загруженных номеров заказов
	GetOrders(ctx context.Context, user string) ([]byte, error)
	// GetBalance - получение текущего баланса пользователя
	GetBalance(ctx context.Context, login string) (float32, float32, error)
	// Debit - запрос на списание средств
	Debit(ctx context.Context, user, order string, sum float32) error
	// GetWithdrawals - получение информации о выводе средств
	GetWithdrawals(ctx context.Context, user string) ([]byte, error)
}

type UseCase struct {
	repo   Repository        // interface Repository
	log    Logger            // interface Logger
	entity *entity.AllEntity // struct entity
	DB     *sql.DB
}

func New(r Repository, l Logger) *UseCase {
	var db *sql.DB
	e := &entity.AllEntity{}
	return &UseCase{repo: r, log: l, entity: e, DB: db}
}
func (uc *UseCase) GetEntity() *entity.AllEntity {
	return uc.entity
}
func (uc *UseCase) Do() *UseCase {
	return uc
}
func (uc *UseCase) DoRegister(ctx context.Context, login, password string, _ *http.Request) error {
	err := uc.repo.Register(ctx, login, password)
	return err
}
func (uc *UseCase) DoAuth(ctx context.Context, login, password string, _ *http.Request) error {
	err := uc.repo.Auth(ctx, login, password)
	return err
}
func (uc *UseCase) DoInsertOrder(ctx context.Context, user string, order string) error {
	err := uc.repo.InsertOrder(ctx, user, order)
	return err
}

func (uc *UseCase) DoGetOrders(ctx context.Context, user string) ([]byte, error) {
	orders, err := uc.repo.GetOrders(ctx, user)
	return orders, err
}

func (uc *UseCase) DoGetBalance(ctx context.Context, login string) (float32, float32, error) {
	b, w, err := uc.repo.GetBalance(ctx, login)
	return b, w, err
}

func (uc *UseCase) DoDebit(ctx context.Context, user, order string, sum float32) error {
	err := uc.repo.Debit(ctx, user, order, sum)
	return err
}

func (uc *UseCase) DoGetWithdrawals(ctx context.Context, user string) ([]byte, error) {
	orders, err := uc.repo.GetWithdrawals(ctx, user)
	return orders, err
}
