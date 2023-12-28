package controller

import "net/http"

type Register struct {
	uc UseCase
}

func NewRegister(uc UseCase) *Register {
	return &Register{uc: uc}
}

func (c *Register) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
