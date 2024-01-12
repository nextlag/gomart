package controller

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
	"github.com/nextlag/gomart/internal/usecase"
)

type Login struct {
	uc  *usecase.UseCase
	log *slog.Logger
}

func NewLogin(uc *usecase.UseCase, log *slog.Logger) *Login {
	return &Login{uc: uc, log: log}
}

func (h *Login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := h.uc.GetEntity().User
	er := usecase.NewErr().GetError()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&user); err != nil {
		h.log.Error("decode JSON", "error Login handler", err.Error())
		http.Error(w, er.DecodeJSON.Error(), http.StatusBadRequest)
		return
	}
	if err := h.uc.DoAuth(r.Context(), user.Login, user.Password, r); err != nil {
		h.log.Error("incorrect login or password", "error Login handler", err.Error())
		http.Error(w, er.Unauthorized.Error(), http.StatusUnauthorized)
		return
	}

	jwtToken, err := auth.SetAuth(user.Login, h.log, w, r)
	if err != nil {
		h.log.Error("can't set cookie", "error Login handler", err.Error())
		http.Error(w, er.NoCookie.Error(), http.StatusInternalServerError)
		return
	}
	l := fmt.Sprintf("[%s] success authenticated", user.Login)
	h.log.Info(l, "token", jwtToken)

	// Возвращаем успешный статус и сообщение об успешной регистрации
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(l))
}
