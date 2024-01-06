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
	login, _ := r.Context().Value(auth.LoginKey).(string)

	// Чтение тела запроса
	body, err := io.ReadAll(r.Body)
	order := string(body)
	switch {
	case order == "":
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	case err != nil:
		h.log.Error("body reading error", "error PostOrder handler", err.Error())
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Проверка формата номера заказа
	ok := luna.CheckValidOrder(order)
	if !ok {
		h.log.Info("invalid order format")
		http.Error(w, "неверный формат заказа", http.StatusUnprocessableEntity)
		return
	}
	err = h.uc.DoInsertOrder(r.Context(), login, order)
	switch {
	case errors.Is(err, usecase.ErrAnotherUser):
		http.Error(w, "номер заказа уже был загружен другим пользователем", http.StatusConflict)
		return
	case errors.Is(err, usecase.ErrThisUser):
		http.Error(w, "номер заказа уже был загружен этим пользователем", http.StatusOK)
		return
	default:
		http.Error(w, "новый номер заказа принят в обработку", http.StatusAccepted)
	}

	// Обработка успешного запроса
	h.log.Info("order received", "login", login, "order", order)
	w.WriteHeader(http.StatusOK)
}
