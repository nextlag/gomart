package controller

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/nextlag/gomart/internal/entity"
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

type OrderResponse struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float32 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

func (h *GetOrders) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value(auth.LoginKey).(string)
	ordersData, err := h.uc.DoGetOrders(r.Context(), user)
	if err != nil {
		h.log.Debug("GetOrders handler", "error", err.Error())
		http.Error(w, h.er.InternalServer.Error(), http.StatusInternalServerError)
		return
	}
	if len(ordersData) == 0 {
		http.Error(w, h.er.NoContent.Error(), http.StatusNoContent)
		return
	}

	// Размаршализовываем JSON в структуру Orders
	var orders []entity.Orders
	if err := json.Unmarshal(ordersData, &orders); err != nil {
		h.log.Error("error unmarshalling json", "GetOrders handler", err.Error())
		http.Error(w, h.er.InternalServer.Error(), http.StatusInternalServerError)
		return
	}

	// Создаем новую структуру для ответа с необходимыми полями
	var responseOrders []OrderResponse
	for _, order := range orders {
		responseOrder := OrderResponse{
			Number:     order.Number,
			Status:     order.Status,
			UploadedAt: order.UploadedAt,
		}

		// Добавить поле Accrual только если оно не нулевое
		if order.Accrual != 0 {
			responseOrder.Accrual = order.Accrual
		}

		responseOrders = append(responseOrders, responseOrder)
	}

	// Маршалинг данных в JSON
	result, err := json.Marshal(responseOrders)
	if err != nil {
		h.log.Error("error when marshaling responseOrders", "GetOrders handler", err.Error())
		http.Error(w, h.er.InternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}
