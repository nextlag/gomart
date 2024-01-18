package controllers

import (
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
)

func (c Controller) Withdrawals(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.LoginKey).(string)
	w.Write([]byte(user))
}
