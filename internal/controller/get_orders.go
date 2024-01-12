package controller

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
	"github.com/nextlag/gomart/internal/usecase"
)

type GetOrders struct {
	uc  *usecase.UseCase
	log *slog.Logger
	er  *usecase.AllErr
}

func NewGetOrders(uc *usecase.UseCase, log *slog.Logger, er *usecase.AllErr) *GetOrders {
	return &GetOrders{uc: uc, log: log, er: er}
}

func (h *GetOrders) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	login, _ := r.Context().Value(auth.LoginKey).(string)
	orders, err := h.uc.DoGetOrders(r.Context(), login)
	if err != nil {
		h.log.Debug("GetOrders", orders, h.er.GetOrders.Error())
		http.Error(w, h.er.GetOrders.Error(), http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		http.Error(w, h.er.NoContent.Error(), http.StatusNoContent)
		return
	}

	jsonResponse, err := json.Marshal(orders)
	if err != nil {
		h.log.Error("error encoding orders to JSON", err)
		http.Error(w, h.er.InternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
