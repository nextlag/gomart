// Package auth - middleware authentication
package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/nextlag/gomart/internal/usecase"
)

// authContextKey - type for the authentication context key
type authContextKey string

// LoginKey - key for the context with the user login
const LoginKey authContextKey = "login"

// CookieAuthentication - checks whether the user is authorized using a cookie.
// If the URL path is not "/api/user/register" or not "/api/user/login" and the user has a valid authorization cookie,
// it serves requests and inserts the login into the context.
// Otherwise, it does not allow execution to continue and returns a 401 status code (if the user is not authenticated)
// or 500 (if an internal server error occurred).
func CookieAuthentication(log usecase.Logger, er *usecase.ErrAll) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			login, err := GetCookie(log, r)

			switch {
			case errors.Is(err, er.ErrToken):
				http.Error(w, er.ErrToken.Error(), http.StatusUnauthorized)
			case errors.Is(err, er.ErrAuth):
				log.Error("error empty login", "error CookieAuthentication", err.Error())
				http.Error(w, er.ErrAuth.Error(), http.StatusUnauthorized)
			case err != nil:
				log.Error("error getting cookie", "error CookieAuthentication", err.Error())
				http.Error(w, er.ErrInternalServer.Error(), http.StatusUnauthorized)
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
