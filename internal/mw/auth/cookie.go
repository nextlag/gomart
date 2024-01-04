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

func WithCookieLogin(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			login, err := GetCookie(log, r)
			if errors.Is(err, errAuth) {
				log.Error("Error getting cookie", "package", "auth", "file", "cookie.go", "error", err.Error())
				http.Error(w, "You are not authenticated", http.StatusUnauthorized)
				return
			} else if err != nil {
				log.Error("Error reading cookie", "package", "auth", "file", "cookie.go", "error", err.Error())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), LoginKey, login)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
