// Package auth - предоставляет функциональность аутентификации пользователя с использованием JSON Web Tokens (JWT) и файлов cookie.
package auth

import (
	"net/http"

	"github.com/golang-jwt/jwt/v4"

	"github.com/nextlag/gomart/internal/config"
	"github.com/nextlag/gomart/internal/usecase"
)

const (
	Cookie = "Auth"
)

// Claims представляет структуру пользовательских клеймов для JWT.
type Claims struct {
	jwt.RegisteredClaims
	Login string `json:"login"`
}

// buildJWTString генерирует токен JWT с предоставленным логином и подписывает его с использованием настроенного секретного ключа.
func buildJWTString(login string, log usecase.Logger) (string, error) {
	// Создает новый токен JWT с пользовательскими клеймами и подписывает его с использованием алгоритма HMAC SHA-256.
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{},
		Login:            login,
	})
	log.Debug("buildJWTString", "config.Cfg.SecretToken", config.Cfg.SecretToken)

	// Подписывает токен с использованием секретного ключа и получает строку токена.
	tokenString, err := jwtToken.SignedString([]byte(config.Cfg.SecretToken))
	if err != nil {
		log.Error("token signing error", "error buildJWTString", err.Error())
		return "", err
	}

	return tokenString, nil
}

// SetAuth создает новую куку для предоставленного логина и устанавливает ее в HTTP-ответе.
func SetAuth(login string, log usecase.Logger, w http.ResponseWriter) (string, error) {
	// Сгенерировать токен JWT для логина.
	jwtToken, err := buildJWTString(login, log)
	if err != nil {
		log.Error("cookie creation error", "error SetAuth", err.Error())
		return "", err
	}

	// Создает новую HTTP-куку с токеном JWT и устанавливает ее в ответе.
	cookie := http.Cookie{
		Name:  Cookie,
		Value: jwtToken,
		Path:  "/",
	}
	http.SetCookie(w, &cookie)
	log.Debug("SetAuth", "received token", jwtToken)

	return jwtToken, nil
}

// getLogin извлекает логин пользователя из предоставленного токена JWT.
// A function used to get a user's login using a JWT. It accepts a JWT and returns a login and error.
func getLogin(tokenString string, log usecase.Logger) (string, error) {
	log.Debug("getLogin", "received token", tokenString)

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Error("unexpected signing method", nil)
			return nil, nil
		}
		return []byte(config.Cfg.SecretToken), nil
	})
	if err != nil {
		log.Error("error parsing token", "getLogin", err.Error())
		return "", usecase.Token
	}
	if !token.Valid {
		log.Error("token is not valid", nil)
		return "", err
	}
	log.Debug("getLogin", "login", claims.Login)
	return claims.Login, nil
}

// GetCookie извлекает логин пользователя из кука "Auth".
func GetCookie(log usecase.Logger, r *http.Request) (string, error) {
	// Извлечь подписанную куку логина из запроса.
	signedLogin, err := r.Cookie(Cookie)
	if err != nil {
		log.Error("error receiving cookie", "error GetCookie", nil)
		return "", usecase.Auth
	}

	// Извлекает логин из токена JWT в куке.
	login, err := getLogin(signedLogin.Value, log)
	if err != nil {
		log.Error("error reading cookie", "error GetCookie", err.Error())
		return "", err
	}
	log.Debug("GetCookie", "login", login)

	return login, nil
}
