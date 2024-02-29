package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/lib/pq"

	"github.com/nextlag/gomart/internal/mw/auth"
	"github.com/nextlag/gomart/pkg/generatestring"
	"github.com/nextlag/gomart/pkg/logger/l"
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
	log := l.L(c.ctx)
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
		log.Error("failed to process the request", l.ErrAttr(err))
		http.Error(w, er.ErrDecodeJSON.Error(), http.StatusBadRequest)
		return
	case len(user.Login) == 0:
		log.Error("error: empty login", "login", user.Login)
		http.Error(w, er.ErrRequest.Error(), http.StatusBadRequest)
		return
	case len(user.Password) == 0:
		user.Password = generatestring.NewRandomString(8)
		log.Info("generating password", "login", user.Login, "password", user.Password)
	}

	log.Debug("findings", "login", user.Login, "password", user.Password)

	// Вызываем метод DoRegister UseCase для выполнения регистрации
	if err = c.uc.DoRegister(c.ctx, user.Login, user.Password, r); err != nil {
		// Обрабатываем ошибку регистрации
		var pqErr *pq.Error
		isPGError := errors.As(err, &pqErr)
		switch {
		case isPGError && pqErr.Code == "23505":
			log.Error("duplicate login", l.ErrAttr(err))
			// Если дубликат логина - возвращаем конфликт
			http.Error(w, er.ErrNoLogin.Error(), http.StatusConflict)
		default:
			log.Error("register error", l.ErrAttr(err))
			// В противном случае возвращаем внутреннюю ошибку сервера
			http.Error(w, er.ErrInternalServer.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Устанавливаем аутентификационную куку после успешной регистрации
	jwt, err := auth.SetAuth(c.ctx, user.Login, w)
	if err != nil {
		log.Error("can't set cookie: ", l.ErrAttr(err))
		http.Error(w, er.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}
	log.Debug("authentication", "login", user.Login, "password", user.Password, "token", jwt)

	// Возвращаем успешный статус и сообщение об успешной регистрации
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
}
