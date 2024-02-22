// Package auth - middleware authentication
package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/nextlag/gomart/internal/usecase"
	"github.com/nextlag/gomart/pkg/logger/l"
)

// authContextKey - ключ контекста аутентификации
type authContextKey string

// LoginKey - ключ для контекста с логином пользователя
const LoginKey authContextKey = "login"

// CookieAuthentication возвращает middleware для аутентификации пользователя по аутентификационной куке.
//
// Эта функция принимает логгер и объект ошибок как параметры и возвращает middleware для аутентификации пользователя.
// Middleware проверяет наличие аутентификационной куки в запросе.
// Если кука отсутствует, возвращает ошибку Unauthorized (401).
// Если происходит ошибка при получении куки, возвращает ошибку Unauthorized (401) или InternalServerError (500),
// в зависимости от характера ошибки.
// Если аутентификационная кука успешно получена, устанавливает логин пользователя в контекст запроса и передает управление следующему обработчику.
//
// Параметры:
//   - log: usecase.Logger - логгер для записи сообщений об ошибках и информации.
//   - er: *usecase.ErrAll - объект, содержащий ошибки, используемые в UseCase.
//
// Возвращаемые значения:
//   - func(http.Handler) http.Handler: middleware для аутентификации пользователя.
func CookieAuthentication(ctx context.Context, er *usecase.ErrAll) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := l.L(ctx)
			// Получаем логин пользователя из аутентификационной куки
			login, err := GetCookie(ctx, r)

			switch {
			case errors.Is(err, er.ErrToken):
				// Если кука не содержит токена, возвращаем ошибку Unauthorized (401)
				http.Error(w, er.ErrToken.Error(), http.StatusUnauthorized)
			case errors.Is(err, er.ErrAuth):
				// Если кука не содержит аутентификационные данные, логируем ошибку и возвращаем ошибку Unauthorized (401)
				log.Error("error empty login", l.ErrAttr(err))
				http.Error(w, er.ErrAuth.Error(), http.StatusUnauthorized)
			case err != nil:
				// Если происходит любая другая ошибка при получении куки, логируем ошибку и возвращаем ошибку Unauthorized (401)
				log.Error("error getting cookie", l.ErrAttr(err))
				http.Error(w, er.ErrInternalServer.Error(), http.StatusUnauthorized)
			default:
				// Создаем новый контекст с установленным логином пользователя
				ctx = context.WithValue(ctx, LoginKey, login)
				// Обновляем запрос с новым контекстом
				r = r.WithContext(ctx)
				log.Debug("CookieAuthentication", "context", ctx.Value(LoginKey), "login", login)

				// Передаем управление следующему обработчику
				next.ServeHTTP(w, r)
			}
		})
	}
}
