package controller

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
)

type Login struct {
	uc  UseCase
	log *slog.Logger
}

func NewLogin(uc UseCase, log *slog.Logger) *Login {
	return &Login{uc: uc, log: log}
}

func (h *Login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var user Credentials
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&user); err != nil {
		http.Error(w, "failed to decode json", http.StatusBadRequest)
		return
	}
	if err := h.uc.DoAuth(r.Context(), user.Login, user.Password, r); err != nil {
		http.Error(w, "incorrect login or password", http.StatusUnauthorized)
		return
	}

	jwtToken, err := auth.SetAuth(user.Login, h.log, w, r)
	if err != nil {
		h.log.Error("can't set cookie", "package", "controller", "files", "authentication.go", "error", err.Error())
		http.Error(w, "can't set cookie", http.StatusInternalServerError)
		return
	}
	l := fmt.Sprintf("[%s] success authenticated", user.Login)
	h.log.Info(l, "token", jwtToken)

	// Возвращаем успешный статус и сообщение об успешной регистрации
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(l))
}
