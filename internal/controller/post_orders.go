package controller

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
	"github.com/nextlag/gomart/internal/usecase"
)

type PostOrders struct {
	uc  *usecase.UseCase
	log *slog.Logger
	er  *usecase.AllErr
}

func NewPostOrders(uc *usecase.UseCase, log *slog.Logger, er *usecase.AllErr) *PostOrders {
	return &PostOrders{uc: uc, log: log, er: er}
}

func (h *PostOrders) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Получаем логин из контекста
	user, _ := r.Context().Value(auth.LoginKey).(string)
	h.log.Debug("get user PostOrders", "user", user)

	// Чтение тела запроса
	body, err := io.ReadAll(r.Body)
	order := string(body)
	switch {
	case order == "":
		http.Error(w, h.er.RequestFormat.Error(), http.StatusBadRequest)
		return
	case err != nil:
		h.log.Error("body reading error", "error PostOrder handler", err.Error())
		http.Error(w, h.er.InternalServer.Error(), http.StatusInternalServerError)
		return
	}

	err = h.uc.DoInsertOrder(r.Context(), user, order)
	switch {
	case errors.Is(err, h.er.OrderFormat):
		h.log.Debug("insert Order 422", "error", err.Error())
		http.Error(w, h.er.OrderFormat.Error(), http.StatusUnprocessableEntity)
		return
	case errors.Is(err, h.er.AnotherUser):
		h.log.Debug("insert Order 409", "error", err.Error())
		http.Error(w, h.er.AnotherUser.Error(), http.StatusConflict)
		return
	case errors.Is(err, h.er.ThisUser):
		h.log.Debug("insert Order 200", "error", err.Error())
		http.Error(w, h.er.ThisUser.Error(), http.StatusOK)
		return
	default:
		http.Error(w, h.er.OrderAccepted.Error(), http.StatusAccepted)
	}

	// Обработка успешного запроса
	h.log.Info("order received", "user", user, "order", order)
	o := fmt.Sprintf("order number: %s\n", order)
	w.Write([]byte(o))
}
