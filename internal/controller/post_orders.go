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
}

func NewPostOrders(uc *usecase.UseCase, log *slog.Logger) *PostOrders {
	return &PostOrders{uc: uc, log: log}
}

func (h *PostOrders) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	er := h.uc.Status()
	// Получаем логин из контекста
	login, _ := r.Context().Value(auth.LoginKey).(string)

	// Чтение тела запроса
	body, err := io.ReadAll(r.Body)
	order := string(body)
	switch {
	case order == "":
		http.Error(w, er.RequestFormat.Error(), http.StatusBadRequest)
		return
	case err != nil:
		h.log.Error("body reading error", "error PostOrder handler", err.Error())
		http.Error(w, er.InternalServer.Error(), http.StatusInternalServerError)
		return
	}

	err = h.uc.DoInsertOrder(r.Context(), login, order)
	switch {
	case errors.Is(err, er.OrderFormat):
		http.Error(w, er.OrderFormat.Error(), http.StatusUnprocessableEntity)
		return
	case errors.Is(err, er.AnotherUser):
		http.Error(w, er.AnotherUser.Error(), http.StatusConflict)
		return
	case errors.Is(err, er.ThisUser):
		http.Error(w, er.ThisUser.Error(), http.StatusOK)
		return
	default:
		http.Error(w, er.OrderAccepted.Error(), http.StatusAccepted)
	}

	// Обработка успешного запроса
	h.log.Info("order received", "login", login, "order", order)
	o := fmt.Sprintf("order number: %s\n", order)
	w.Write([]byte(o))
}
