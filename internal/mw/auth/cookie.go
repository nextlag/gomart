package auth

import (
	"context"
	"log/slog"
	"net/http"
)

// AuthContextKey - тип для ключа контекста аутентификации
type AuthContextKey string

// LoginKey - ключ для контекста с логином пользователя
const LoginKey AuthContextKey = "login"

// WithCookieLogin создает промежуточное программное обеспечение для проверки аутентификации пользователя с использованием файлов cookie.
func WithCookieLogin(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Извлекает логин пользователя из файла cookie.
			userLogin, err := GetCookie(r, log)
			if err == errAuth {
				// Если пользователь не аутентифицирован, вернуть код состояния 401 и сообщение об ошибке.
				log.Error("error receiving cookie", "error auth|cookie.go", err.Error())
				http.Error(w, "you are not authenticated", http.StatusUnauthorized)
				return
			} else if err != nil {
				// Если произошла внутренняя ошибка при чтении файла cookie, вернуть код состояния 500.
				log.Error("error reading cookie", "error auth|cookie.go", err.Error())
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			// Добавляет логин в контекст запроса.
			ctx := context.WithValue(r.Context(), LoginKey, userLogin)
			r = r.WithContext(ctx)

			// Передает управление следующему обработчику в цепочке.
			next.ServeHTTP(w, r)
		})
	}
}
