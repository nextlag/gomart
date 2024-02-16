package controllers

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
)

// PostOrders - designed to process HTTP POST requests containing information about orders.
// It accepts data from the client, inserts the order into the database and returns the
// appropriate HTTP status depending on the result of the operation.
func (c *Controller) PostOrders(w http.ResponseWriter, r *http.Request) {
	er := c.uc.Do().Err()
	user, _ := r.Context().Value(auth.LoginKey).(string)
	c.log.Debug("get user PostOrders", "user", user)

	body, err := io.ReadAll(r.Body)
	order := string(body)
	switch {
	case order == "":
		http.Error(w, er.ErrRequestFormat.Error(), http.StatusBadRequest)
		return
	case err != nil:
		c.log.Error("body reading error", "error PostOrder handler", err.Error())
		http.Error(w, er.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	err = c.uc.DoInsertOrder(r.Context(), user, order)
	switch {
	case errors.Is(err, er.ErrOrderFormat):
		c.log.Error("insert Order 422", "error", err.Error())
		http.Error(w, er.ErrOrderFormat.Error(), http.StatusUnprocessableEntity)
		return
	case errors.Is(err, er.ErrAnotherUser):
		c.log.Error("insert Order 409", "error", err.Error())
		http.Error(w, er.ErrAnotherUser.Error(), http.StatusConflict)
		return
	case errors.Is(err, er.ErrThisUser):
		c.log.Error("insert Order 200", "error", err.Error())
		http.Error(w, er.ErrThisUser.Error(), http.StatusOK)
		return
	default:
		http.Error(w, er.ErrOrderAccepted.Error(), http.StatusAccepted)
	}

	c.log.Info("order received", "user", user, "order", order)
	o := fmt.Sprintf("order number: %s\n", order)
	w.Write([]byte(o))
}
