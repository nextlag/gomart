package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
)

func (c Controller) Authentication(w http.ResponseWriter, r *http.Request) {
	user := c.uc.Do().GetEntity()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&user); err != nil {
		c.log.Error("decode JSON", "error Login handler", err.Error())
		http.Error(w, c.er.DecodeJSON.Error(), http.StatusBadRequest)
		return
	}
	if err := c.uc.DoAuth(r.Context(), user.Login, user.Password, r); err != nil {
		c.log.Error("incorrect login or password", "error Login handler", err.Error())
		http.Error(w, c.er.Unauthorized.Error(), http.StatusUnauthorized)
		return
	}

	jwtToken, err := auth.SetAuth(user.Login, c.log, w)
	if err != nil {
		c.log.Error("can't set cookie", "error Login handler", err.Error())
		http.Error(w, c.er.NoCookie.Error(), http.StatusInternalServerError)
		return
	}
	l := fmt.Sprintf("[%s] success authenticated", user.Login)
	c.log.Info(l, "token", jwtToken)

	// Возвращаем успешный статус и сообщение об успешной регистрации
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(l))
}
