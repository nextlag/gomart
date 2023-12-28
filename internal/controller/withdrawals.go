package controller

import "net/http"

type Withdrawals struct {
	uc UseCase
}

func NewWithdrawals(uc UseCase) *Withdrawals {
	return &Withdrawals{uc: uc}
}

func (c *Withdrawals) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
