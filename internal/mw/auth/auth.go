// Пакет auth предоставляет функциональность аутентификации пользователя с использованием JSON Web Tokens (JWT) и файлов cookie.

package auth

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v4"

	"github.com/nextlag/gomart/internal/config"
)

// Claims представляет структуру пользовательских клеймов для JWT.
type Claims struct {
	jwt.RegisteredClaims
	login string
}

// errAuth - ошибка, указывающая, что пользователь не аутентифицирован.
var errAuth = errors.New("вы не аутентифицированы")

// buildJWTString генерирует токен JWT с предоставленным логином и подписывает его с использованием настроенного секретного ключа.
func buildJWTString(login string, log *slog.Logger) (string, error) {
	// Создает новый токен JWT с пользовательскими клеймами и подписывает его с использованием алгоритма HMAC SHA-256.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{},
		login:            login,
	})

	// Подписывает токен с использованием секретного ключа и получает строку токена.
	tokenString, err := token.SignedString([]byte(config.Cfg.SecretToken))
	if err != nil {
		log.Error("Ошибка подписи токена: ", err)
		return "", err
	}

	return tokenString, nil
}

// SetAuth создает новую аутентификационную куку для предоставленного логина и устанавливает ее в HTTP-ответе.
func SetAuth(res http.ResponseWriter, login string, log *slog.Logger) error {
	// Сгенерировать токен JWT для логина.
	jwt, err := buildJWTString(login, log)
	if err != nil {
		log.Error("Ошибка создания куки: ", err)
		return err
	}

	// Создает новую HTTP-куку с токеном JWT и устанавливает ее в ответе.
	cookie := http.Cookie{
		Name:  "Auth",
		Value: jwt,
		Path:  "/",
	}
	http.SetCookie(res, &cookie)

	return nil
}

// getLogin извлекает логин пользователя из предоставленного токена JWT.
func getLogin(tokenString string, log *slog.Logger) (string, error) {
	// Парсит токен JWT и извлекает клеймы.
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		// Проверяет метод подписи.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Error("Непредвиденный метод подписи", nil)
			return nil, nil
		}
		// Возвращает секретный ключ для валидации.
		return []byte(config.Cfg.SecretToken), nil
	})
	if err != nil {
		return "", err
	}
	// Проверяет, является ли токен действительным.
	if !token.Valid {
		log.Error("Токен не действителен", nil)
		return "", err
	}
	return claims.login, nil
}

// GetCookie извлекает логин пользователя из куки "Auth" в предоставленном HTTP-запросе.
func GetCookie(req *http.Request, log *slog.Logger) (string, error) {
	// Извлечь подписанную куку логина из запроса.
	signedLogin, err := req.Cookie("Auth")
	if err != nil {
		log.Error("Ошибка при получении куки", err)
		return "", errAuth
	}

	// Извлекает логин из токена JWT в куке.
	login, err := getLogin(signedLogin.Value, log)
	if err != nil {
		log.Error("Ошибка при чтении куки", err)
		return "", err
	}

	return login, nil
}
