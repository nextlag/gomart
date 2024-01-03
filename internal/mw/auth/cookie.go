package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
)

// authContextKey - тип для ключа контекста аутентификации
type authContextKey string

// LoginKey - ключ для контекста с логином пользователя
const LoginKey authContextKey = "login"

// WithCookieLogin - проверка аутентификации пользователя с использованием файлов cookie.
func WithCookieLogin(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Извлекает логин пользователя из файла cookie.
			userLogin, err := GetCookie(log, r)
			if errors.Is(err, errAuth) {
				// Если пользователь не аутентифицирован, вернуть 40Ω
				log.Error("error receiving cookie", "error auth|cookie.go", err.Error())
				http.Error(w, "you are not authenticated", http.StatusUnauthorized)
				return
			} else if err != nil {
				// Если произошла ошибка при чтении файла cookie, вернуть 500.
				log.Error("error reading cookie", "error auth|cookie.go", err.Error())
				http.Error(w, "error reading cookie", http.StatusInternalServerError)
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
