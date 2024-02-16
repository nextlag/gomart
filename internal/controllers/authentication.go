package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
)

// Authentication processes the user authentication request.
// It extracts user data from the request body, calls the DoAuth method from the use case
// to check the correctness of the login and password, and in case of successful authentication sets
// authentication cookie and returns a successful status.
func (c *Controller) Authentication(w http.ResponseWriter, r *http.Request) {
	user := c.uc.Do().GetEntity()
	er := c.uc.Do().Err()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&user); err != nil {
		c.log.Error("decode JSON", "error Login handler", err.Error())
		http.Error(w, er.ErrDecodeJSON.Error(), http.StatusBadRequest)
		return
	}
	if err := c.uc.DoAuth(r.Context(), user.Login, user.Password, r); err != nil {
		c.log.Error("incorrect login or password", "error Login handler", err.Error())
		http.Error(w, er.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	jwtToken, err := auth.SetAuth(user.Login, c.log, w)
	if err != nil {
		c.log.Error("can't set cookie", "error Login handler", err.Error())
		http.Error(w, er.ErrNoCookie.Error(), http.StatusInternalServerError)
		return
	}
	l := fmt.Sprintf("[%s] success authenticated", user.Login)
	c.log.Info(l, "token", jwtToken)

	// Возвращаем успешный статус и сообщение об успешной регистрации
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(l))
}
