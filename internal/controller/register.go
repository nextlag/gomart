package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/lib/pq"

	"github.com/nextlag/gomart/internal/entity"
	"github.com/nextlag/gomart/internal/mw/auth"
	"github.com/nextlag/gomart/internal/usecase"
	"github.com/nextlag/gomart/pkg/generatestring"
)

// Register представляет собой контроллер для обработки регистрации пользователя.
type Register struct {
	uc     UseCase // UseCase для обработки бизнес-логики регистрации
	log    *slog.Logger
	er     *usecase.ErrStatus
	entity *entity.Entity
}

// NewRegister создает новый экземпляр контроллера Register.
func NewRegister(uc UseCase, log *slog.Logger, er *usecase.ErrStatus, entity *entity.Entity) *Register {
	return &Register{uc: uc, log: log, er: er, entity: entity}
}

// ServeHTTP обрабатывает HTTP-запросы для регистрации пользователя.
func (h *Register) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Декодируем JSON-данные из тела запроса в структуру Credentials
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&h.entity.User); err != nil {
		http.Error(w, h.er.DecodeJSON.Error(), http.StatusBadRequest)
		return
	}

	// Проверяем обязательные поля
	switch {
	case len(h.entity.User.Login) == 0:
		h.log.Info("error: empty login")
		http.Error(w, h.er.Request.Error(), http.StatusBadRequest)
		return
	case len(h.entity.User.Password) == 0:
		h.log.Info("generating password")
		h.entity.User.Password = generatestring.NewRandomString(8)
	}

	h.log.Debug("authorization", "login", h.entity.User.Login, "password", h.entity.User.Password)

	// Вызываем метод DoRegister UseCase для выполнения регистрации
	if err := h.uc.DoRegister(r.Context(), h.entity.User.Login, h.entity.User.Password, r); err != nil {
		// Обрабатываем ошибку регистрации
		var pqErr *pq.Error
		isPGError := errors.As(err, &pqErr)
		switch {
		case isPGError && pqErr.Code == "23505":
			// Если дубликат логина - возвращаем конфликт
			http.Error(w, "login is already token", http.StatusConflict)
		default:
			// В противном случае возвращаем внутреннюю ошибку сервера
			http.Error(w, h.er.InternalServer.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Устанавливаем аутентификационную куку после успешной регистрации
	jwt, err := auth.SetAuth(h.entity.User.Login, h.log, w, r)
	if err != nil {
		h.log.Error("can't set cookie: ", "error controller|register", err.Error())
		http.Error(w, h.er.InternalServer.Error(), http.StatusInternalServerError)
		return
	}
	h.log.Debug("authentication", "login", h.entity.User.Login, "token", jwt)

	// Возвращаем успешный статус и сообщение об успешной регистрации
	w.WriteHeader(http.StatusOK)
	cookie := fmt.Sprintf("Cookie: %s=%s\n", auth.Cookie, jwt)
	w.Write([]byte(cookie))
}
