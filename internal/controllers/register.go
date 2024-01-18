package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/lib/pq"

	"github.com/nextlag/gomart/internal/mw/auth"
	"github.com/nextlag/gomart/pkg/generatestring"
)

// Register обрабатывает HTTP-запросы для регистрации пользователя.
func (c Controller) Register(w http.ResponseWriter, r *http.Request) {
	user := c.uc.Do().GetEntity()
	er := c.uc.Do().Er()
	// Декодируем JSON-данные из тела запроса в структуру Credentials
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&user); err != nil {
		c.log.Error("Decode JSON", "Login", user.Login, "error Register handler", err.Error())
		http.Error(w, er.DecodeJSON.Error(), http.StatusBadRequest)
		return
	}

	// Проверяем обязательные поля
	switch {
	case len(user.Login) == 0:
		c.log.Info("error: empty login")
		http.Error(w, er.Request.Error(), http.StatusBadRequest)
		return
	case len(user.Password) == 0:
		c.log.Info("generating password")
		user.Password = generatestring.NewRandomString(8)
	}

	c.log.Debug("authorization", "login", user.Login, "password", user.Password)

	// Вызываем метод DoRegister UseCase для выполнения регистрации
	if err := c.uc.DoRegister(r.Context(), user.Login, user.Password, r); err != nil {
		// Обрабатываем ошибку регистрации
		var pqErr *pq.Error
		isPGError := errors.As(err, &pqErr)
		switch {
		case isPGError && pqErr.Code == "23505":
			// Если дубликат логина - возвращаем конфликт
			http.Error(w, er.NoLogin.Error(), http.StatusConflict)
		default:
			// В противном случае возвращаем внутреннюю ошибку сервера
			http.Error(w, er.InternalServer.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Устанавливаем аутентификационную куку после успешной регистрации
	jwt, err := auth.SetAuth(user.Login, c.log, w)
	if err != nil {
		c.log.Error("can't set cookie: ", "error controllers|register", err.Error())
		http.Error(w, er.InternalServer.Error(), http.StatusInternalServerError)
		return
	}
	c.log.Debug("authentication", "login", user.Login, "token", jwt)

	// Возвращаем успешный статус и сообщение об успешной регистрации
	w.WriteHeader(http.StatusOK)
	cookie := fmt.Sprintf("Cookie: %s=%s\n", auth.Cookie, jwt)
	w.Write([]byte(cookie))
}
