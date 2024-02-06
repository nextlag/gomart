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
func (c *Controller) Register(w http.ResponseWriter, r *http.Request) {
	uc := c.uc.Do()
	user := uc.GetEntity()
	er := uc.Err()
	// Декодируем JSON-данные из тела запроса в структуру Credentials
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&user)
	defer func() {
		if panicErr := recover(); panicErr != nil {
			c.log.Error("panic during JSON decoding", "error", panicErr, "body", r.Body)
			http.Error(w, er.ErrDecodeJSON.Error(), http.StatusBadRequest)
		}
	}()

	switch {
	case err != nil:
		c.log.Error("failed to process the request", "error", err.Error())
		http.Error(w, er.ErrDecodeJSON.Error(), http.StatusBadRequest)
		return
	case len(user.Login) == 0:
		c.log.Error("error: empty login", "error", err, "login", user.Login)
		http.Error(w, er.ErrRequest.Error(), http.StatusBadRequest)
		return
	case len(user.Password) == 0:
		user.Password = generatestring.NewRandomString(8)
		c.log.Info("generating password", "login", user.Login, "password", user.Password)
	}

	c.log.Debug("findings", "login", user.Login, "password", user.Password)

	// Вызываем метод DoRegister UseCase для выполнения регистрации
	if err := uc.Register(r.Context(), user.Login, user.Password); err != nil {
		// Обрабатываем ошибку регистрации
		var pqErr *pq.Error
		isPGError := errors.As(err, &pqErr)
		switch {
		case isPGError && pqErr.Code == "23505":
			c.log.Error("Duplicate login", "error", err.Error())
			// Если дубликат логина - возвращаем конфликт
			http.Error(w, er.ErrNoLogin.Error(), http.StatusConflict)
		default:
			// В противном случае возвращаем внутреннюю ошибку сервера
			http.Error(w, er.ErrInternalServer.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Устанавливаем аутентификационную куку после успешной регистрации
	jwt, err := auth.SetAuth(user.Login, c.log, w)
	if err != nil {
		c.log.Error("can't set cookie: ", "error controllers|register", err.Error())
		http.Error(w, er.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}
	c.log.Debug("authentication", "login", user.Login, "password", user.Password, "token", jwt)

	// Возвращаем успешный статус и сообщение об успешной регистрации
	w.WriteHeader(http.StatusOK)
	cookie := fmt.Sprintf("Cookie: %s=%s\n", auth.Cookie, jwt)
	w.Write([]byte(cookie))
}
