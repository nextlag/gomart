package controller

import (
	"log/slog"
	"net/http"
)

type PostOrders struct {
	uc UseCase
}

func NewPostOrders(uc UseCase, log *slog.Logger) *PostOrders {
	return &PostOrders{uc: uc}
}

func (c *PostOrders) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
