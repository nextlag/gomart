package controller

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
	"github.com/nextlag/gomart/internal/usecase"
	"github.com/nextlag/gomart/pkg/luna"
)

type PostOrders struct {
	uc  UseCase
	log *slog.Logger
}

func NewPostOrders(uc UseCase, log *slog.Logger) *PostOrders {
	return &PostOrders{uc: uc, log: log}
}

func (h *PostOrders) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Получаем логин из контекста
	login, ok := r.Context().Value(auth.LoginKey).(string)
	if !ok {
		h.log.Error("error getting login from context", "package", "controller", "file", "post_orders.go")
		http.Error(w, "you are not authenticated", http.StatusUnauthorized)
		return
	}

	// Чтение тела запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.Error("body reading error", "error PostOrder handler", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Проверка формата номера заказа
	order := string(body)
	ok = luna.CheckValidOrder(order)
	if !ok {
		h.log.Info("invalid order format")
		http.Error(w, "invalid order format", http.StatusUnprocessableEntity)
		return
	}
	err = h.uc.DoInsertOrder(r.Context(), login, order)
	switch {
	case errors.Is(err, usecase.ErrAlreadyLoadedOrder):
		http.Error(w, "the order number has already been uploaded by another user", http.StatusConflict)
		return
	case errors.Is(err, usecase.ErrYouAlreadyLoadedOrder):
		http.Error(w, "the order number has already been uploaded by this user", http.StatusOK)
		return
	default:
		http.Error(w, "successfully loaded order", http.StatusAccepted)
	}

	// Обработка успешного запроса
	h.log.Info("order received", "login", login, "order", order)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("order successfully processed"))
}
