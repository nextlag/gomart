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
		http.Error(w, usecase.ErrRequestFormat.Error(), http.StatusBadRequest)
		return
	case err != nil:
		h.log.Error("body reading error", "error PostOrder handler", err.Error())
		http.Error(w, usecase.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	err = h.uc.DoInsertOrder(r.Context(), login, order)
	switch {
	case errors.Is(err, usecase.ErrOrderFormat):
		http.Error(w, usecase.ErrOrderFormat.Error(), http.StatusUnprocessableEntity)
		return
	case errors.Is(err, usecase.ErrAnotherUser):
		http.Error(w, usecase.ErrAnotherUser.Error(), http.StatusConflict)
		return
	case errors.Is(err, usecase.ErrThisUser):
		http.Error(w, usecase.ErrThisUser.Error(), http.StatusOK)
		return
	default:
		http.Error(w, usecase.ErrOrderAccepted.Error(), http.StatusAccepted)
	}

	// Обработка успешного запроса
	h.log.Info("order received", "login", login, "order", order)
	o := fmt.Sprintf("order number: %s\n", order)
	w.Write([]byte(o))
}
