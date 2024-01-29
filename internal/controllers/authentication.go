package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
	"github.com/nextlag/gomart/pkg/logger/l"
)

// Authentication обрабатывает запрос на аутентификацию пользователя.
//
// Этот метод принимает запрос HTTP POST, содержащий JSON с данными пользователя.
// При успешной аутентификации метод устанавливает аутентификационный токен в куки и возвращает
// статус OK (200) с сообщением об успешной аутентификации.
// В случае, если происходит ошибка при декодировании JSON, метод возвращает статус BadRequest (400)
// с соответствующим сообщением об ошибке.
// Если логин или пароль пользователя неверны, метод возвращает статус Unauthorized (401)
// с сообщением об ошибке аутентификации.
// При любых других ошибках метод возвращает статус InternalServerError (500)
// с сообщением об ошибке.
//
// Параметры:
//   - w: http.ResponseWriter - объект для записи HTTP-ответа.
//   - r: *http.Request - объект HTTP-запроса.
//
// Возвращаемые значения:
//   - нет.
func (c *Controller) Authentication(w http.ResponseWriter, r *http.Request) {
	log := l.L(c.ctx)
	// Получаем данные пользователя из UseCase
	user := c.uc.Do().GetEntity()
	// Получаем объект ошибки из UseCase
	er := c.uc.Do().Err()

	// Декодируем JSON из тела запроса в объект пользователя
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&user); err != nil {
		// Если произошла ошибка при декодировании JSON, возвращаем ошибку BadRequest
		log.Error("decode JSON", l.ErrAttr(err))
		http.Error(w, er.ErrDecodeJSON.Error(), http.StatusBadRequest)
		return
	}

	// Проверяем логин и пароль пользователя
	if err := c.uc.DoAuth(c.ctx, user.Login, user.Password, r); err != nil {
		// Если логин или пароль неверны, возвращаем ошибку Unauthorized
		log.Error("incorrect login or password", l.ErrAttr(err))
		http.Error(w, er.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	// Устанавливаем аутентификационный токен в куки
	jwtToken, err := auth.SetAuth(c.ctx, user.Login, w)
	if err != nil {
		// Если не удалось установить куки, возвращаем ошибку InternalServerError
		log.Error("can't set cookie", l.ErrAttr(err))
		http.Error(w, er.ErrNoCookie.Error(), http.StatusInternalServerError)
		return
	}
	// Логируем успешную аутентификацию
	l := fmt.Sprintf("[%s] success authenticated", user.Login)
	log.Debug(l, "token", jwtToken)

	// Возвращаем успешный статус и сообщение об успешной аутентификации
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(l))
}
