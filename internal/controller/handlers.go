package controller

import (
	"context"
	"net/http"
)

type UseCase interface {
	DoRegister(ctx context.Context, login, password string) error
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

func New(uc UseCase) *Handlers {
	return &Handlers{
		Balance:     NewBalance(uc).ServeHTTP,
		GetOrders:   NewGetOrders(uc).ServeHTTP,
		Login:       NewLogin(uc).ServeHTTP,
		PostOrders:  NewPostOrders(uc).ServeHTTP,
		Register:    NewRegister(uc).ServeHTTP,
		Withdraw:    NewWithdraw(uc).ServeHTTP,
		Withdrawals: NewWithdrawals(uc).ServeHTTP,
	}
}
