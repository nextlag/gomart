package controllers

import (
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
)

// GetOrders - designed to process HTTP GET requests aimed at obtaining a list of user orders.
// It returns a list of orders in JSON format.
func (c *Controller) GetOrders(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value(auth.LoginKey).(string)
	er := c.uc.Do().Err()
	result, err := c.uc.DoGetOrders(r.Context(), user)
	if err != nil {
		c.log.Error("handler GetOrders", "error", err.Error())
		http.Error(w, er.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}
