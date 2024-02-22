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

// Register обрабатывает запрос на регистрацию нового пользователя.
//
// Этот метод принимает запрос HTTP POST с JSON-данными, содержащими логин и пароль нового пользователя.
// При успешной регистрации метод устанавливает аутентификационную куку и возвращает статус OK (200).
// Если происходит ошибка при декодировании JSON или при обработке запроса, метод возвращает ошибку BadRequest (400)
// с соответствующим сообщением об ошибке.
// Если указанный логин уже занят другим пользователем, метод возвращает ошибку Conflict (409).
// В случае любых других ошибок при регистрации, метод возвращает ошибку InternalServerError (500).
//
// Параметры:
//   - w: http.ResponseWriter - объект для записи HTTP-ответа.
//   - r: *http.Request - объект HTTP-запроса.
//
// Возвращаемые значения:
//   - нет.
func (c *Controller) Register(w http.ResponseWriter, r *http.Request) {
	// Получаем объект пользователя из UseCase
	user := c.uc.Do().GetEntity()
	// Получаем объект ошибки из UseCase
	er := c.uc.Do().Err()

	// Декодируем JSON-данные из тела запроса в структуру Credentials
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&user)
	// Проверяем наличие ошибок при декодировании
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
	if err := c.uc.DoRegister(r.Context(), user.Login, user.Password, r); err != nil {
		// Обрабатываем ошибку регистрации
		var pqErr *pq.Error
		isPGError := errors.As(err, &pqErr)
		switch {
		case isPGError && pqErr.Code == "23505":
			c.log.Error("Duplicate login", "error", err.Error())
			// Если дубликат логина - возвращаем конфликт
			http.Error(w, er.ErrNoLogin.Error(), http.StatusConflict)
		default:
			c.log.Error("Register error", "error", err.Error())
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
