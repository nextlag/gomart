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

// Balance processes the request to get the user's balance.
// It retrieves the user's login from the request context, calls the DoGetBalance method from the use case,
// to get the current balance and amount of user debits, and returns this data in JSON format.
func (c *Controller) Balance(w http.ResponseWriter, r *http.Request) {
	login, _ := r.Context().Value(auth.LoginKey).(string)
	balance, withdrawn, err := c.uc.DoGetBalance(r.Context(), login)
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
	c.log.Info("GetBalance handler", "balance", balance, "withdrawn", withdrawn)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}
