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

// WithCookieLogin - проверяет, авторизован ли пользователь с использованием cookie.
// Если URL-путь не "/api/user/register" или не "/api/user/login" и у пользователя есть действительная кука авторизации,
// она обслуживает запросы и вставляет логин в контекст.
// В противном случае она не позволяет продолжить выполнение и возвращает статус кода 401 (если пользователь не аутентифицирован),
// или 500 (если произошла внутренняя ошибка сервера).
func WithCookieLogin(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			login, err := GetCookie(log, r)
			switch {
			case errors.Is(err, errAuth):
				log.Error("error getting cookie", "package", "auth", "file", "cookie.go", "error", err.Error())
				http.Error(w, "you are not authenticated", http.StatusUnauthorized)
				return
			case err != nil:
				log.Error("error reading cookie", "package", "auth", "file", "cookie.go", "error", err.Error())
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			case login == "":
				log.Error("empty login", "package", "auth", "file", "cookie.go")
				http.Error(w, "you are not authenticated", http.StatusUnauthorized)
				return
			}

			// Создаем новый контекст с установленным логином
			ctx := context.WithValue(r.Context(), LoginKey, login)
			// Обновляем запрос с новым контекстом
			r = r.WithContext(ctx)
			log.Debug("WithCookieLogin", "context", ctx.Value(LoginKey), "login", login)

			next.ServeHTTP(w, r)
		})
	}
}
