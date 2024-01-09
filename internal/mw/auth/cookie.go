package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/nextlag/gomart/internal/usecase"
)

// authContextKey - тип для ключа контекста аутентификации
type authContextKey string

// LoginKey - ключ для контекста с логином пользователя
const LoginKey authContextKey = "login"

// CookieAuthentication - проверяет, авторизован ли пользователь с использованием cookie.
// Если URL-путь не "/api/user/register" или не "/api/user/login" и у пользователя есть действительная кука авторизации,
// она обслуживает запросы и вставляет логин в контекст.
// В противном случае она не позволяет продолжить выполнение и возвращает статус кода 401 (если пользователь не аутентифицирован),
// или 500 (если произошла внутренняя ошибка сервера).
func CookieAuthentication(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			login, err := GetCookie(log, r)

			switch {
			case errors.Is(err, usecase.Status().Token):
				http.Error(w, usecase.Status().Token.Error(), http.StatusUnauthorized)
			case errors.Is(err, usecase.Status().Auth):
				log.Error("error empty login", "error CookieAuthentication", err.Error())
				http.Error(w, usecase.Status().Auth.Error(), http.StatusUnauthorized)
			case err != nil:
				log.Error("error getting cookie", "error CookieAuthentication", err.Error())
				http.Error(w, usecase.Status().InternalServer.Error(), http.StatusUnauthorized)
			default:
				// Создаем новый контекст с установленным логином
				ctx := context.WithValue(r.Context(), LoginKey, login)
				// Обновляем запрос с новым контекстом
				r = r.WithContext(ctx)
				log.Debug("CookieAuthentication", "context", ctx.Value(LoginKey), "login", login)

				next.ServeHTTP(w, r)
			}
		})
	}
}
