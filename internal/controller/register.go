package controller

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/lib/pq"

	"github.com/nextlag/gomart/internal/mw/auth"
)

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
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&userData); err != nil {
		http.Error(w, "Wrong request", http.StatusBadRequest)
		return
	}

	if userData.Login == "" || userData.Password == "" {
		http.Error(w, "Wrong request", http.StatusBadRequest)
		return
	}

	err := h.uc.DoRegister(r.Context(), userData.Login, userData.Password)
	if err != nil {
		pqErr, isPGError := err.(*pq.Error)
		if isPGError && pqErr.Code == "23505" {
			http.Error(w, "Login is already taken", http.StatusConflict)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	err = auth.SetAuth(w, userData.Login, h.log)
	if err != nil {
		h.log.Error("Can't set cookie: ", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully registered"))
}
