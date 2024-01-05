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
		Path:  r.URL.Path,
	}
	http.SetCookie(w, &cookie)

	return jwtToken, nil
}

// getLogin извлекает логин пользователя из предоставленного токена JWT.
// A function used to get a user's login using a JWT. It accepts a JWT and returns a login and error.
func getLogin(tokenString string, log *slog.Logger) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Error("unexpected signing method", nil)
			return nil, nil
		}
		return []byte(config.Cfg.SecretToken), nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		log.Error("Token is not valid", nil)
		return "", err
	}
	return claims.Login, nil
}

// GetCookie извлекает логин пользователя из кука "Auth".
func GetCookie(log *slog.Logger, r *http.Request) (string, error) {
	signedLogin, err := r.Cookie("Auth")
	if err != nil {
		log.Error("Error getting cookie", "error", err.Error())
		return "", errAuth
	}

	login, err := getLogin(signedLogin.Value, log)
	if err != nil {
		log.Error("Error reading cookie", "error", err.Error())
		return "", err
	}

	return login, nil
}
