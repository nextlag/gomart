package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
)

type userBalance struct {
	Balance   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

func (c Controller) Balance(w http.ResponseWriter, r *http.Request) {
	login, _ := r.Context().Value(auth.LoginKey).(string)
	balance, withdrawn, err := c.uc.DoGetBalance(r.Context(), login)
	c.log.Debug("GetBalance handler", "balance", balance, "withdrawn", withdrawn)
	if err != nil {
		c.log.Error("Balance handler", "balance", balance, "withdrawn", withdrawn, "error", err.Error())
		http.Error(w, "error get balance", http.StatusInternalServerError)
		return
	}
	user := userBalance{
		Balance:   balance,
		Withdrawn: withdrawn,
	}
	result, err := json.Marshal(user)
	if err != nil {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}
