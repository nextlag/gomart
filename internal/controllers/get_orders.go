package controllers

import (
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
)

func (c *Controller) GetOrders(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value(auth.LoginKey).(string)
	uc := c.uc.Do()
	er := uc.Err()
	result, err := uc.GetOrders(r.Context(), user)
	if err != nil {
		c.log.Info("handler GetOrders", "error", err.Error())
		http.Error(w, er.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}
