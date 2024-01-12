package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/lib/pq"

	"github.com/nextlag/gomart/internal/mw/auth"
	"github.com/nextlag/gomart/internal/usecase"
	"github.com/nextlag/gomart/pkg/generatestring"
)

// Register представляет собой контроллер для обработки регистрации пользователя.
type Register struct {
	uc  *usecase.UseCase // UseCase для обработки бизнес-логики регистрации
	log *slog.Logger
}

// NewRegister создает новый экземпляр контроллера Register.
func NewRegister(uc *usecase.UseCase, log *slog.Logger) *Register {
	return &Register{uc: uc, log: log}
}

// ServeHTTP обрабатывает HTTP-запросы для регистрации пользователя.
func (h *Register) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := h.uc.GetEntity().User
	er := usecase.NewErr().GetError()
	// Декодируем JSON-данные из тела запроса в структуру Credentials
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&user); err != nil {
		h.log.Error("Decode JSON", "Login", user.Login, "error Register handler", err.Error())
		http.Error(w, er.DecodeJSON.Error(), http.StatusBadRequest)
		return
	}

	// Проверяем обязательные поля
	switch {
	case len(user.Login) == 0:
		h.log.Info("error: empty login")
		http.Error(w, er.Request.Error(), http.StatusBadRequest)
		return
	case len(user.Password) == 0:
		h.log.Info("generating password")
		user.Password = generatestring.NewRandomString(8)
	}

	h.log.Debug("authorization", "login", user.Login, "password", user.Password)

	// Вызываем метод DoRegister UseCase для выполнения регистрации
	if err := h.uc.DoRegister(r.Context(), user.Login, user.Password, r); err != nil {
		// Обрабатываем ошибку регистрации
		var pqErr *pq.Error
		isPGError := errors.As(err, &pqErr)
		switch {
		case isPGError && pqErr.Code == "23505":
			// Если дубликат логина - возвращаем конфликт
			http.Error(w, "login is already token", http.StatusConflict)
		default:
			// В противном случае возвращаем внутреннюю ошибку сервера
			http.Error(w, er.InternalServer.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Устанавливаем аутентификационную куку после успешной регистрации
	jwt, err := auth.SetAuth(user.Login, h.log, w, r)
	if err != nil {
		h.log.Error("can't set cookie: ", "error controller|register", err.Error())
		http.Error(w, er.InternalServer.Error(), http.StatusInternalServerError)
		return
	}
	h.log.Debug("authentication", "login", user.Login, "token", jwt)

	// Возвращаем успешный статус и сообщение об успешной регистрации
	w.WriteHeader(http.StatusOK)
	cookie := fmt.Sprintf("Cookie: %s=%s\n", auth.Cookie, jwt)
	w.Write([]byte(cookie))
}
