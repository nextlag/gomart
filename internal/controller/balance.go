package controller

import (
	"log/slog"
	"net/http"
)

type Balance struct {
	uc  UseCase
	log *slog.Logger
}

func NewBalance(uc UseCase, log *slog.Logger) *Balance {
	return &Balance{uc: uc, log: log}
}

func (c *Balance) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
