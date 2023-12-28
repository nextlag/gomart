package controller

import "net/http"

type Withdraw struct {
	uc UseCase
}

func NewWithdraw(uc UseCase) *Withdraw {
	return &Withdraw{uc: uc}
}

func (c *Withdraw) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
