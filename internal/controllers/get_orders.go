package controllers

import (
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
)

func (c Controller) GetOrders(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value(auth.LoginKey).(string)
	result, err := c.uc.DoGetOrders(r.Context(), user)
	if err != nil {
		c.log.Debug("Обработчик GetOrders", "ошибка", err.Error())
		http.Error(w, c.er.InternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}
