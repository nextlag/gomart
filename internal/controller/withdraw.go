package controller

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
	"github.com/nextlag/gomart/internal/usecase"
)

type Withdraw struct {
	uc  UseCase
	log *slog.Logger
	er  *usecase.AllErr
}

func NewWithdraw(uc UseCase, log *slog.Logger, er *usecase.AllErr) *Withdraw {
	return &Withdraw{uc: uc, log: log, er: er}
}

// A struct used to parse a json request to withdraw bonuses making an order.
type debet struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}

func (h *Withdraw) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Получаем логин из контекста
	user, _ := r.Context().Value(auth.LoginKey).(string)
	var request debet
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}
	h.log.Debug("debet request", "user", user, "order", request.Order, "sum", request.Sum)
	err := h.uc.DoDebit(r.Context(), user, request.Order, request.Sum)
	switch {
	case errors.Is(err, h.er.NoBalance):
		h.log.Info("на счету недостаточно средств", "NoBalance", h.er.NoBalance.Error())
		http.Error(w, h.er.NoBalance.Error(), http.StatusPaymentRequired)
		return
	case errors.Is(err, h.er.OrderFormat):
		http.Error(w, h.er.OrderFormat.Error(), http.StatusUnprocessableEntity)
		return
	case errors.Is(err, h.er.ThisUser) || errors.Is(err, h.er.AnotherUser):
		http.Error(w, "order is already loaded", http.StatusUnprocessableEntity)
		return
	case err != nil:
		http.Error(w, h.er.InternalServer.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("bonuses were written off success"))
}
