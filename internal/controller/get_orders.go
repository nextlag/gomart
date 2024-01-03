package controller

import (
	"log/slog"
	"net/http"
)

type GetOrders struct {
	uc UseCase
}

func NewGetOrders(uc UseCase, log *slog.Logger) *GetOrders {
	return &GetOrders{uc: uc}
}

func (c *GetOrders) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
