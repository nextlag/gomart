// Package auth - provides user authentication functionality using JSON Web Tokens (JWT) and cookies.
package auth

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v4"

	"github.com/nextlag/gomart/internal/config"
	"github.com/nextlag/gomart/internal/usecase"
	"github.com/nextlag/gomart/pkg/logger/l"
)

const (
	Cookie = "ErrAuth"
)

// Claims introduces a custom claims framework for JWT.
type Claims struct {
	jwt.RegisteredClaims
	Login string `json:"login"`
}

// buildJWTString generates a JWT token with the provided login and signs it using the configured secret key.
func buildJWTString(ctx context.Context, login string) (string, error) {
	log := l.L(ctx)
	// Создает новый токен JWT с пользовательскими клеймами и подписывает его с использованием алгоритма HMAC SHA-256.
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{},
		Login:            login,
	})
	log.Debug("buildJWTString", "config.Cfg.SecretToken", config.Cfg.SecretToken)

	// Подписывает токен с использованием секретного ключа и получает строку токена.
	tokenString, err := jwtToken.SignedString([]byte(config.Cfg.SecretToken))
	if err != nil {
		log.Error("token signing error", l.ErrAttr(err))
		return "", err
	}

	return tokenString, nil
}

// SetAuth creates a new cookie for the provided login and sets it in the HTTP response.
func SetAuth(ctx context.Context, login string, w http.ResponseWriter) (string, error) {
	log := l.L(ctx)
	// Сгенерировать токен JWT для логина.
	jwtToken, err := buildJWTString(ctx, login)
	if err != nil {
		log.Error("cookie creation error", l.ErrAttr(err))
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
func getLogin(ctx context.Context, tokenString string) (string, error) {
	log := l.L(ctx)
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
		log.Error("error parsing token", l.ErrAttr(err))
		return "", usecase.ErrToken
	}
	if !token.Valid {
		log.Error("token is not valid", nil)
		return "", err
	}
	log.Debug("getLogin", "login", claims.Login)
	return claims.Login, nil
}

// GetCookie retrieves the user's login from the "ErrAuth" cookie.
func GetCookie(ctx context.Context, r *http.Request) (string, error) {
	log := l.L(ctx)
	// Извлечь подписанную куку логина из запроса.
	signedLogin, err := r.Cookie(Cookie)
	if err != nil {
		log.Error("error receiving cookie", "error GetCookie", nil)
		return "", usecase.ErrAuth
	}

	// Извлекает логин из токена JWT в куке.
	login, err := getLogin(ctx, signedLogin.Value)
	if err != nil {
		log.Error("error reading cookie", "error GetCookie", err.Error())
		return "", err
	}
	log.Debug("GetCookie", "login", login)

	return login, nil
}
