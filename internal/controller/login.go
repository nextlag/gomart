package controller

import (
	"encoding/json"
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
	if err := h.uc.DoRegister(r.Context(), user.Login, user.Password, r); err != nil {
		http.Error(w, "incorrect login or password", http.StatusUnauthorized)
		return
	}

	_, err := auth.SetAuth(user.Login, h.log, w, r)
	if err != nil {
		h.log.Error("can't set cookie", "error controller|register", err.Error())
		http.Error(w, "can't set cookie", http.StatusInternalServerError)
		return
	}

	// Возвращаем успешный статус и сообщение об успешной регистрации
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("you have successfully logged in"))
}
