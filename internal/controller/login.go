package controller

import (
	"log/slog"
	"net/http"
)

type Login struct {
	uc UseCase
}

func NewLogin(uc UseCase, log *slog.Logger) *Login {
	return &Login{uc: uc}
}

func (c *Login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
