package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
)

// A struct used to parse a json request to withdraw bonuses making an order.
type debit struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}

func (c *Controller) Withdraw(w http.ResponseWriter, r *http.Request) {
	uc := c.uc.Do()
	er := uc.Err()
	// Получаем логин из контекста
	user, _ := r.Context().Value(auth.LoginKey).(string)
	var request debit
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}
	c.log.Debug("debet request", "user", user, "order", request.Order, "sum", request.Sum)
	err := uc.Debit(r.Context(), user, request.Order, request.Sum)
	switch {
	case errors.Is(err, er.ErrNoBalance):
		c.log.Error("there are insufficient funds in the account", "NoBalance", er.ErrNoBalance.Error())
		http.Error(w, er.ErrNoBalance.Error(), http.StatusPaymentRequired)
		return
	case errors.Is(err, er.ErrOrderFormat):
		c.log.Error("Withdraw OrderFormat", "error", err.Error())
		http.Error(w, er.ErrOrderFormat.Error(), http.StatusUnprocessableEntity)
		return
	case errors.Is(err, er.ErrThisUser) || errors.Is(err, er.ErrAnotherUser):
		c.log.Debug("withdraw", "user", user, "order", request.Order)
		c.log.Error("Withdraw AnotherUser", "error", err.Error())
		http.Error(w, "order is already loaded", http.StatusConflict)
		return
	case err != nil:
		c.log.Error("Withdraw handler", "error", err.Error())
		http.Error(w, er.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("bonuses were written off success"))
}
