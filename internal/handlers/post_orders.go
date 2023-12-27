package handlers

import "net/http"

type PostOrders struct {
	uc UseCase
}

func NewPostOrders(uc UseCase) *PostOrders {
	return &PostOrders{uc: uc}
}

func (c *PostOrders) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
