package controller

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/nextlag/gomart/internal/entity"
	"github.com/nextlag/gomart/internal/mw/auth"
	"github.com/nextlag/gomart/internal/usecase"
)

type Login struct {
	uc     UseCase
	log    *slog.Logger
	er     *usecase.ErrStatus
	entity *entity.Entity
}

func NewLogin(uc UseCase, log *slog.Logger, er *usecase.ErrStatus, entity *entity.Entity) *Login {
	return &Login{uc: uc, log: log, er: er, entity: entity}
}

func (h *Login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&h.entity.User); err != nil {
		http.Error(w, "failed to decode json", http.StatusBadRequest)
		return
	}
	if err := h.uc.DoAuth(r.Context(), h.entity.User.Login, h.entity.User.Password, r); err != nil {
		http.Error(w, "incorrect login or password", http.StatusUnauthorized)
		return
	}

	jwtToken, err := auth.SetAuth(h.entity.User.Login, h.log, w, r)
	if err != nil {
		h.log.Error("can't set cookie", "error Login handler", err.Error())
		http.Error(w, "can't set cookie", http.StatusInternalServerError)
		return
	}
	l := fmt.Sprintf("[%s] success authenticated", h.entity.User.Login)
	h.log.Info(l, "token", jwtToken)

	// Возвращаем успешный статус и сообщение об успешной регистрации
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(l))
}
