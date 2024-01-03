package controller

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/lib/pq"

	"github.com/nextlag/gomart/internal/mw/auth"
)

// Register представляет собой контроллер для обработки регистрации пользователя.
type Register struct {
	uc  UseCase      // UseCase для обработки бизнес-логики регистрации
	log *slog.Logger // Логгер для записи логов
}

// NewRegister создает новый экземпляр контроллера Register.
func NewRegister(uc UseCase) *Register {
	return &Register{uc: uc}
}

// ServeHTTP обрабатывает HTTP-запросы для регистрации пользователя.
func (h *Register) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var user Credentials

	// Декодируем JSON-данные из тела запроса в структуру Credentials
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&user); err != nil {
		http.Error(w, "Wrong request", http.StatusBadRequest)
		return
	}

	// Проверяем обязательные поля
	if user.Login == "" || user.Password == "" {
		http.Error(w, "Wrong request", http.StatusBadRequest)
		return
	}

	// Вызываем метод DoRegister UseCase для выполнения регистрации
	if err := h.uc.DoRegister(r.Context(), user.Login, user.Password); err != nil {
		// Обрабатываем ошибку регистрации
		pqErr, isPGError := err.(*pq.Error)
		switch {
		case isPGError && pqErr.Code == "23505":
			// Если ошибка уникального нарушения, возвращаем конфликт
			http.Error(w, "Login is already taken", http.StatusConflict)
		default:
			// В противном случае возвращаем внутреннюю ошибку сервера
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Устанавливаем аутентификационную куку после успешной регистрации
	if err := auth.SetAuth(w, user.Login, h.log); err != nil {
		h.log.Error("Can't set cookie: ", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Возвращаем успешный статус и сообщение об успешной регистрации
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully registered"))
}
