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
	case errors.Is(err, c.er.NoBalance):
		c.log.Info("на счету недостаточно средств", "NoBalance", c.er.NoBalance.Error())
		http.Error(w, c.er.NoBalance.Error(), http.StatusPaymentRequired)
		return
	case errors.Is(err, c.er.OrderFormat):
		http.Error(w, c.er.OrderFormat.Error(), http.StatusUnprocessableEntity)
		return
	case errors.Is(err, c.er.ThisUser) || errors.Is(err, c.er.AnotherUser):
		http.Error(w, "order is already loaded", http.StatusUnprocessableEntity)
		return
	case err != nil:
		http.Error(w, c.er.InternalServer.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("bonuses were written off success"))
}
