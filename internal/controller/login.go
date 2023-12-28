package controller

import "net/http"

type Login struct {
	uc UseCase
}

func NewLogin(uc UseCase) *Login {
	return &Login{uc: uc}
}

func (c *Login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
