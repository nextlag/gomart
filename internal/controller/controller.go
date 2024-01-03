package controller

import (
	"context"
	"log/slog"
	"net/http"
)

type UseCase interface {
	DoRegister(ctx context.Context, login, password string) error
}

// A struct used to get and store data from a json requests.
type Credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
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

func New(uc UseCase, log *slog.Logger) *Handlers {
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
