package controllers

import (
	"errors"
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
)

// Withdrawals - designed to process HTTP GET requests aimed at obtaining information about bonus
// debits from the user's account. It returns a list of bonus write-off transactions in JSON format.
func (c *Controller) Withdrawals(w http.ResponseWriter, r *http.Request) {
	er := c.uc.Do().Err()
	// Получаем логин из контекста
	user, _ := r.Context().Value(auth.LoginKey).(string)

	result, err := c.uc.DoGetWithdrawals(r.Context(), user)
	switch {
	case errors.Is(err, er.ErrNoRows):
		c.log.Error("withdrawals handler", "error no rows", err.Error())
		http.Error(w, "there is no write off", http.StatusNoContent)
		return
	case err != nil:
		c.log.Error("withdrawals handler", "error", err.Error())
		http.Error(w, er.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}
