package controller

import (
	"log/slog"
	"net/http"
)

type Withdraw struct {
	uc UseCase
}

func NewWithdraw(uc UseCase, log *slog.Logger) *Withdraw {
	return &Withdraw{uc: uc}
}

func (c *Withdraw) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
