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

func (c Controller) Withdraw(w http.ResponseWriter, r *http.Request) {
	er := c.uc.Do().Er()
	// Получаем логин из контекста
	user, _ := r.Context().Value(auth.LoginKey).(string)
	var request debit
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}
	c.log.Debug("debet request", "user", user, "order", request.Order, "sum", request.Sum)
	err := c.uc.Do().DoDebit(r.Context(), user, request.Order, request.Sum)
	switch {
	case errors.Is(err, er.NoBalance):
		c.log.Error("there are insufficient funds in the account", "NoBalance", er.NoBalance.Error())
		http.Error(w, er.NoBalance.Error(), http.StatusPaymentRequired)
		return
	case errors.Is(err, er.OrderFormat):
		c.log.Error("Withdraw OrderFormat", "error", err.Error())
		http.Error(w, er.OrderFormat.Error(), http.StatusUnprocessableEntity)
		return
	case errors.Is(err, er.AnotherUser):
		c.log.Debug("withdraw", "user", user, "order", request.Order)
		c.log.Error("Withdraw AnotherUser", "error", err.Error())
		http.Error(w, "order is already loaded", http.StatusUnprocessableEntity)
		return
	case err != nil:
		c.log.Error("Withdraw handler", "error", err.Error())
		http.Error(w, er.InternalServer.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("bonuses were written off success"))
}
