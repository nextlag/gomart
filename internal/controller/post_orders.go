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
	uc  UseCase
	log *slog.Logger
	er  *usecase.ErrStatus
}

func NewPostOrders(uc UseCase, log *slog.Logger, er *usecase.ErrStatus) *PostOrders {
	return &PostOrders{uc: uc, log: log, er: er}
}

func (h *PostOrders) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Получаем логин из контекста
	login, _ := r.Context().Value(auth.LoginKey).(string)

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

	err = h.uc.DoInsertOrder(r.Context(), login, order)
	switch {
	case errors.Is(err, h.er.OrderFormat):
		http.Error(w, h.er.OrderFormat.Error(), http.StatusUnprocessableEntity)
		return
	case errors.Is(err, h.er.AnotherUser):
		http.Error(w, h.er.AnotherUser.Error(), http.StatusConflict)
		return
	case errors.Is(err, h.er.ThisUser):
		http.Error(w, h.er.ThisUser.Error(), http.StatusOK)
		return
	default:
		http.Error(w, h.er.OrderAccepted.Error(), http.StatusAccepted)
	}

	// Обработка успешного запроса
	h.log.Info("order received", "login", login, "order", order)
	o := fmt.Sprintf("order number: %s\n", order)
	w.Write([]byte(o))
}
