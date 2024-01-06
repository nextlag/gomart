package controller

import (
	"log/slog"
	"net/http"
)

type Withdrawals struct {
	uc UseCase
}

func NewWithdrawals(uc UseCase, log *slog.Logger) *Withdrawals {
	return &Withdrawals{uc: uc}
}

func (c *Withdrawals) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
