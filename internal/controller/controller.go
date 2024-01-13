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
	DoInsertOrder(ctx context.Context, user, order string) error
	DoGetOrders(ctx context.Context, user string) ([]byte, error)
	DoGetBalance(ctx context.Context, login string) (float32, float32, error)
	DoDebit(ctx context.Context, user, numOrder string, sum float32) error
}

type Handlers struct {
	Authentication http.HandlerFunc
	Balance        http.HandlerFunc
	GetOrders      http.HandlerFunc
	PostOrders     http.HandlerFunc
	Register       http.HandlerFunc
	Withdraw       http.HandlerFunc
	Withdrawals    http.HandlerFunc
}

func New(uc *usecase.UseCase, log *slog.Logger, er *usecase.AllErr) *Handlers {
	return &Handlers{
		Authentication: NewLogin(uc, log, er).ServeHTTP,
		Balance:        NewBalance(uc, log, er).ServeHTTP,
		GetOrders:      NewGetOrders(uc, log, er).ServeHTTP,
		PostOrders:     NewPostOrders(uc, log, er).ServeHTTP,
		Register:       NewRegister(uc, log, er).ServeHTTP,
		Withdraw:       NewWithdraw(uc, log, er).ServeHTTP,
		Withdrawals:    NewWithdrawals(uc, log, er).ServeHTTP,
	}
}

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type userBalance struct {
	Balance   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}
