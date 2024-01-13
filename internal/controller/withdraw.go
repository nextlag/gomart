package controller

import (
	"errors"
	"log/slog"
	"net/http"

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
	user, _ := r.Context().Value("login").(string)
	var request debet

	err := h.uc.DoDebit(r.Context(), user, request.Order, request.Sum)
	switch {
	case errors.Is(err, h.er.NoBalance):
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
