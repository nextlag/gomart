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
			case errors.Is(err, errToken):
				http.Error(w, "неверная сигнатура токена", http.StatusUnauthorized)
			case errors.Is(err, errAuth):
				log.Error("error empty login", "error CookieAuthentication", err.Error())
				http.Error(w, "ошибка аутентификации: не задан логин", http.StatusUnauthorized)
				return
			case errors.Is(err, errAuth):
				log.Error("error getting cookie", "error CookieAuthentication", err.Error())
				http.Error(w, "ошибка аутентификации", http.StatusUnauthorized)
				return
			case err != nil:
				log.Error("error reading cookie", "error CookieAuthentication", err.Error())
				http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
				return
			}

			// Создаем новый контекст с установленным логином
			ctx := context.WithValue(r.Context(), LoginKey, login)
			// Обновляем запрос с новым контекстом
			r = r.WithContext(ctx)
			log.Debug("CookieAuthentication", "context", ctx.Value(LoginKey), "login", login)

			next.ServeHTTP(w, r)
		})
	}
}
