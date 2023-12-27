package handlers

import "net/http"

type GetOrders struct {
	uc UseCase
}

func NewGetOrders(uc UseCase) *GetOrders {
	return &GetOrders{uc: uc}
}

func (c *GetOrders) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
