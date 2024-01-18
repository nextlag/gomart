package controllers

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
)

func (c Controller) PostOrders(w http.ResponseWriter, r *http.Request) {
	er := c.uc.Do().Er()
	// Получаем логин из контекста
	user, _ := r.Context().Value(auth.LoginKey).(string)
	c.log.Debug("get user PostOrders", "user", user)

	// Чтение тела запроса
	body, err := io.ReadAll(r.Body)
	order := string(body)
	switch {
	case order == "":
		http.Error(w, er.RequestFormat.Error(), http.StatusBadRequest)
		return
	case err != nil:
		c.log.Error("body reading error", "error PostOrder handler", err.Error())
		http.Error(w, er.InternalServer.Error(), http.StatusInternalServerError)
		return
	}

	err = c.uc.DoInsertOrder(r.Context(), user, order)
	switch {
	case errors.Is(err, er.OrderFormat):
		c.log.Debug("insert Order 422", "error", err.Error())
		http.Error(w, er.OrderFormat.Error(), http.StatusUnprocessableEntity)
		return
	case errors.Is(err, er.AnotherUser):
		c.log.Debug("insert Order 409", "error", err.Error())
		http.Error(w, er.AnotherUser.Error(), http.StatusConflict)
		return
	case errors.Is(err, er.ThisUser):
		c.log.Debug("insert Order 200", "error", err.Error())
		http.Error(w, er.ThisUser.Error(), http.StatusOK)
		return
	default:
		http.Error(w, er.OrderAccepted.Error(), http.StatusAccepted)
	}

	// Обработка успешного запроса
	c.log.Info("order received", "user", user, "order", order)
	o := fmt.Sprintf("order number: %s\n", order)
	w.Write([]byte(o))
}
