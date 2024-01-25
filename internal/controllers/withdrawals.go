package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
)

func (c Controller) Withdrawals(w http.ResponseWriter, r *http.Request) {
	er := c.uc.Do().Err()
	// Получаем логин из контекста
	user, _ := r.Context().Value(auth.LoginKey).(string)

	withdrawals, err := c.uc.DoGetWithdrawals(r.Context(), user)
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

	result, err := json.Marshal(withdrawals)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}
