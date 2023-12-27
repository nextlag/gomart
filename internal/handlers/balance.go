package handlers

import "net/http"

type Balance struct {
	uc UseCase
}

func NewBalance(uc UseCase) *Balance {
	return &Balance{uc: uc}
}

func (c *Balance) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
