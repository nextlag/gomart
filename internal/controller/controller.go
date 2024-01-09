package controller

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/nextlag/gomart/internal/usecase"
)

type UseCase interface {
	DoRegister(ctx context.Context, login, password string, r *http.Request) error
	DoAuth(ctx context.Context, login, password string, r *http.Request) error
	DoInsertOrder(ctx context.Context, login, order string) error
	DoGetOrders(ctx context.Context, login string) (*[]usecase.UseCase, error)
}

type User struct {
	Login    string `json:"login"`
	Password string `json:"passwowrd"`
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

func New(uc *usecase.UseCase, log *slog.Logger) *Handlers {
	return &Handlers{
		Balance:     NewBalance(uc, log).ServeHTTP,
		GetOrders:   NewGetOrders(uc, log).ServeHTTP,
		Login:       NewLogin(uc, log).ServeHTTP,
		PostOrders:  NewPostOrders(uc, log).ServeHTTP,
		Register:    NewRegister(uc, log).ServeHTTP,
		Withdraw:    NewWithdraw(uc, log).ServeHTTP,
		Withdrawals: NewWithdrawals(uc, log).ServeHTTP,
	}
}
