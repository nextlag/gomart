package controller

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
	"github.com/nextlag/gomart/internal/usecase"
)

type Balance struct {
	uc  UseCase
	log *slog.Logger
	er  *usecase.AllErr
}

func NewBalance(uc UseCase, log *slog.Logger, er *usecase.AllErr) *Balance {
	return &Balance{uc: uc, log: log, er: er}
}

func (h *Balance) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	login, _ := r.Context().Value(auth.LoginKey).(string)
	balance, withdrawn, err := h.uc.DoGetBalance(r.Context(), login)
	h.log.Debug("GetBalance handler", "balance", balance, "withdrawn", withdrawn)
	if err != nil {
		h.log.Error("GetBalance handler", "balance", balance, "withdrawn", withdrawn, "error", err.Error())
		http.Error(w, "error get balance", http.StatusInternalServerError)
		return
	}
	user := userBalance{
		Balance:   balance,
		Withdrawn: withdrawn,
	}
	result, err := json.Marshal(user)
	if err != nil {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}
