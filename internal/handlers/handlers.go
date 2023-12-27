package handlers

import "net/http"

type UseCase interface {
	DoRequest()
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
