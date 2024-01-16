package controller

import (
	"log/slog"
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
	"github.com/nextlag/gomart/internal/usecase"
)

type GetOrders struct {
	uc  UseCase
	log *slog.Logger
	er  *usecase.AllErr
}

func NewGetOrders(uc UseCase, log *slog.Logger, er *usecase.AllErr) *GetOrders {
	return &GetOrders{uc: uc, log: log, er: er}
}

func (h *GetOrders) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value(auth.LoginKey).(string)
	result, err := h.uc.DoGetOrders(r.Context(), user)
	if err != nil {
		h.log.Debug("Обработчик GetOrders", "ошибка", err.Error())
		http.Error(w, h.er.InternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}
