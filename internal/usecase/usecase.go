package usecase

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/nextlag/gomart/internal/config"
	"github.com/nextlag/gomart/internal/entity"
)

//go:generate mockgen -destination=mocks.go -package=usecase github.com/nextlag/gomart/internal/usecase Repository
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
	repo   Repository // interface Repository
	cfg    config.HTTPServer
	entity *entity.AllEntity // struct entity
	DB     *sql.DB
}

func New(r Repository, cfg config.HTTPServer) *UseCase {
	e := &entity.AllEntity{}
	return &UseCase{repo: r, cfg: cfg, entity: e}
}

// GetEntity - method returning entity structures
func (uc *UseCase) GetEntity() *entity.AllEntity {
	return uc.entity
}

// UseCase interface methods

func (uc *UseCase) Do() *UseCase {
	return uc
}
func (uc *UseCase) DoRegister(ctx context.Context, login, password string, _ *http.Request) error {
	err := uc.repo.Register(ctx, login, password)
	return err
}
func (uc *UseCase) DoAuth(ctx context.Context, login, password string, _ *http.Request) error {
	return uc.repo.Auth(ctx, login, password)
}
func (uc *UseCase) DoInsertOrder(ctx context.Context, user string, order string) error {
	return uc.repo.InsertOrder(ctx, user, order)
}

func (uc *UseCase) DoGetOrders(ctx context.Context, user string) ([]byte, error) {
	return uc.repo.GetOrders(ctx, user)
}

func (uc *UseCase) DoGetBalance(ctx context.Context, login string) (float32, float32, error) {
	return uc.repo.GetBalance(ctx, login)
}

func (uc *UseCase) DoDebit(ctx context.Context, user, order string, sum float32) error {
	return uc.repo.Debit(ctx, user, order, sum)
}

func (uc *UseCase) DoGetWithdrawals(ctx context.Context, user string) ([]byte, error) {
	return uc.repo.GetWithdrawals(ctx, user)
}
