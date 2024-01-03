package controller

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
)

// HTTP коды состояния
const (
	StatusBadRequest          = http.StatusBadRequest
	StatusInternalServerError = http.StatusInternalServerError
	StatusOK                  = http.StatusOK
)

// ErrValidation представляет ошибку валидации данных.
var ErrValidation = errors.New("validation error")

type Register struct {
	uc  UseCase
	log *slog.Logger
}

func NewRegister(uc UseCase) *Register {
	return &Register{uc: uc}
}

func (h *Register) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var userData Credentials
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&userData); err != nil {
		sendError(w, StatusBadRequest, "Wrong request")
		return
	}

	if err := h.validateAndRegister(w, r, userData); err != nil {
		sendError(w, StatusInternalServerError, "Internal server error")
		return
	}

	if err := auth.SetAuth(w, userData.Login, h.log); err != nil {
		h.log.Error("Can't set cookie: ", err)
		sendError(w, StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(StatusOK)
	w.Write([]byte("Successfully registered"))
}

func (h *Register) validateAndRegister(w http.ResponseWriter, r *http.Request, userData Credentials) error {
	if userData.Login == "" || userData.Password == "" {
		sendError(w, StatusBadRequest, "Wrong request")
		return ErrValidation
	}

	if err := h.uc.DoRegister(r.Context(), userData.Login, userData.Password); err != nil {
		sendError(w, StatusBadRequest, err.Error())
		return err
	}

	return nil
}

func sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(`{"error": "` + message + `"}`))
}
