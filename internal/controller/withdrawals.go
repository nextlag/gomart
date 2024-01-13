package controller

import (
	"log/slog"
	"net/http"

	"github.com/nextlag/gomart/internal/usecase"
)

type Withdrawals struct {
	uc UseCase
}

func NewWithdrawals(uc UseCase, log *slog.Logger, er *usecase.AllErr) *Withdrawals {
	return &Withdrawals{uc: uc}
}

func (c *Withdrawals) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
