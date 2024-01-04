package controller

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
	"github.com/nextlag/gomart/internal/usecase"
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
	login, _ := r.Context().Value(auth.LoginKey).(string)

	// Чтение тела запроса
	body, err := io.ReadAll(r.Body)
	order := string(body)
	switch {
	case order == "":
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	case err != nil:
		h.log.Error("body reading error", "error PostOrder handler", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = h.uc.DoInsertOrder(r.Context(), login, order)
	switch {
	case errors.Is(err, usecase.ErrOrderFormat):
		http.Error(w, "incorrect order format", http.StatusUnprocessableEntity)
		return
	case errors.Is(err, usecase.ErrAnotherUser):
		http.Error(w, "the order number has already been uploaded by another user", http.StatusConflict)
		return
	case errors.Is(err, usecase.ErrThisUser):
		http.Error(w, "the order number has already been uploaded by this user", http.StatusOK)
		return
	default:
		http.Error(w, "new order number accepted for processing", http.StatusAccepted)
	}

	// Обработка успешного запроса
	h.log.Info("order received", "login", login, "order", order)
	w.WriteHeader(http.StatusOK)
}
