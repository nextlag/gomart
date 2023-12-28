// Пакет auth предоставляет промежуточное программное обеспечение для аутентификации пользователя с использованием файлов cookie.

package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
)

// WithCookieLogin создает промежуточное программное обеспечение для проверки аутентификации пользователя с использованием файлов cookie.
// Если URL-путь не "/api/user/register" или "/api/user/login" и у пользователя есть действительная кука для входа в систему,
// оно обслуживает HTTPS-запросы и вставляет логин в контекст.
// В противном случае оно возвращает код состояния 401, если пользователь не аутентифицирован, или 500 в случае внутренней ошибки сервера.
func WithCookieLogin(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// Извлекает логин пользователя из файла cookie.
			userLogin, err := GetCookie(r, log)
			if errors.Is(err, errAuth) {
				// Если пользователь не аутентифицирован, вернуть код состояния 401 и сообщение об ошибке.
				log.Error("Ошибка получения файла cookie", err)
				http.Error(w, "Вы не аутентифицированы", http.StatusUnauthorized)
				return
			} else if err != nil {
				// Если произошла внутренняя ошибка при чтении файла cookie, вернуть код состояния 500.
				log.Error("Ошибка чтения файла cookie", err)
				http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
				return
			}

			// Добавляет логин в контекст запроса.
			ctx := r.Context()
			ctx = context.WithValue(ctx, "login", userLogin)
			r = r.WithContext(ctx)

			// Передает управление следующему обработчику в цепочке.
			next.ServeHTTP(w, r)
		}
		// Оборачивает функцию в тип http.Handler.
		return http.HandlerFunc(fn)
	}
}
