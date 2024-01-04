// Package auth - предоставляет функциональность аутентификации пользователя с использованием JSON Web Tokens (JWT) и файлов cookie.
package auth

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v4"

	"github.com/nextlag/gomart/internal/config"
)

const (
	Cookie = "Auth"
)

// Claims представляет структуру пользовательских клеймов для JWT.
type Claims struct {
	jwt.RegisteredClaims
	Login string `json:"login"`
}

// errAuth - ошибка, указывающая, что пользователь не аутентифицирован.
var errAuth = errors.New("вы не аутентифицированы")

// buildJWTString генерирует токен JWT с предоставленным логином и подписывает его с использованием настроенного секретного ключа.
func buildJWTString(login string, log *slog.Logger) (string, error) {
	// Создает новый токен JWT с пользовательскими клеймами и подписывает его с использованием алгоритма HMAC SHA-256.
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{},
		Login:            login,
	})

	// Подписывает токен с использованием секретного ключа и получает строку токена.
	tokenString, err := jwtToken.SignedString([]byte(config.Cfg.SecretToken))
	if err != nil {
		log.Error("token signing error", "error auth", err.Error())
		return "", err
	}

	return tokenString, nil
}

// SetAuth создает новую куку для предоставленного логина и устанавливает ее в HTTP-ответе.
func SetAuth(login string, log *slog.Logger, w http.ResponseWriter, r *http.Request) (string, error) {
	// Сгенерировать токен JWT для логина.
	jwtToken, err := buildJWTString(login, log)
	if err != nil {
		log.Error("cookie creation error", "error auth", err.Error())
		return "", err
	}

	// Создает новую HTTP-куку с токеном JWT и устанавливает ее в ответе.
	cookie := http.Cookie{
		Name:  "Auth",
		Value: jwtToken,
		Path:  "/",
		// Path:  r.URL.Path,
	}
	http.SetCookie(w, &cookie)

	return jwtToken, nil
}

// getLogin извлекает логин пользователя из предоставленного токена JWT.
func getLogin(tokenString string, log *slog.Logger) (string, error) {
	// Парсит токен JWT и извлекает клеймы.
	claims := &Claims{}
	jwtToken, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		// Проверяет метод подписи.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Error("unexpected signature method", nil)
			return nil, nil
		}
		// Возвращает секретный ключ для валидации.
		return []byte(config.Cfg.SecretToken), nil
	})
	if err != nil {
		return "", err
	}
	// Проверяет, является ли токен действительным.
	if !jwtToken.Valid {
		log.Error("token invalid", "error auth", nil)
		return "", err
	}
	return claims.Login, nil
}

// GetCookie извлекает логин пользователя из кука "Auth".
func GetCookie(log *slog.Logger, r *http.Request) (string, error) {
	// Извлечь подписанную куку логина из запроса.
	signedLogin, err := r.Cookie(Cookie)
	if err != nil {
		log.Error("error receiving cookie", "error auth", nil)
		return "", errAuth
	}

	// Извлекает логин из токена JWT в куке.
	login, err := getLogin(signedLogin.Value, log)
	if err != nil {
		log.Error("error reading cookie", "error auth", err.Error())
		return "", err
	}

	return login, nil
}
