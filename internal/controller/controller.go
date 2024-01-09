package controller

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/nextlag/gomart/internal/entity"
	"github.com/nextlag/gomart/internal/usecase"
)

type UseCase interface {
	DoRegister(ctx context.Context, login, password string, r *http.Request) error
	DoAuth(ctx context.Context, login, password string, r *http.Request) error
	DoInsertOrder(ctx context.Context, login string, order string) error
}

type Handlers struct {
	Balance     http.HandlerFunc
	GetOrders   http.HandlerFunc
	Login       http.HandlerFunc
	PostOrders  http.HandlerFunc
	Register    http.HandlerFunc
	Withdraw    http.HandlerFunc
	Withdrawals http.HandlerFunc
}

func New(uc UseCase, log *slog.Logger, er *usecase.ErrStatus, entity *entity.Entity) *Handlers {
	return &Handlers{
		Balance:     NewBalance(uc, log).ServeHTTP,
		GetOrders:   NewGetOrders(uc, log).ServeHTTP,
		Login:       NewLogin(uc, log, er, entity).ServeHTTP,
		PostOrders:  NewPostOrders(uc, log, er).ServeHTTP,
		Register:    NewRegister(uc, log, er, entity).ServeHTTP,
		Withdraw:    NewWithdraw(uc, log).ServeHTTP,
		Withdrawals: NewWithdrawals(uc, log).ServeHTTP,
	}
}
