package usecase

import (
	"errors"
)

var (
	// ErrNoBalance - недостаточно баланса.
	ErrNoBalance = errors.New("not enough balance")
	// ErrNoRows - строки не найдены
	ErrNoRows = errors.New("no rows were found")
)

// ErrAuth - ошибка, указывающая, что пользователь не аутентифицирован.
var ()

type ErrorAuth struct {
	Auth  error // Auth ошибка аутентификации
	Token error // Token неверная сигнатура токена
}

type ErrCommon struct {
	InternalServer error // InternalServer - status 500: внутренняя ошибка сервера
}

type ErrRegistration struct {
	Request    error
	DecodeJSON error
}

type ErrPostOrder struct {
	ThisUser      error // ThisUser status 200: номер заказа уже был загружен этим пользователем
	AnotherUser   error // AnotherUser status 409: номер заказа уже был загружен другим пользователем
	OrderAccepted error // OrderAccepted status 202: новый номер заказа принят в обработку
	RequestFormat error // RequestFormat status 400: неверный формат запроса
	UnauthUser    error // UnauthUser status 401: пользователь не аутентифицирован
	OrderFormat   error // OrderFormat status 422: неверный формат номера заказа
}

type ErrStatus struct {
	*ErrorAuth
	*ErrCommon
	*ErrRegistration
	*ErrPostOrder
}

func Status() *ErrStatus {
	return &ErrStatus{
		&ErrorAuth{
			Auth:  errors.New("authentication error"),
			Token: errors.New("signature is invalid"),
		},
		&ErrCommon{
			InternalServer: errors.New("insert server error"),
		},
		&ErrRegistration{
			Request:    errors.New("error request"),
			DecodeJSON: errors.New("failed to decode json"),
		},
		&ErrPostOrder{
			ThisUser:      errors.New("the order number has already been uploaded by this user"),
			AnotherUser:   errors.New("the order number has already been uploaded by another user"),
			OrderAccepted: errors.New("new order number accepted for processing"),
			RequestFormat: errors.New("invalid request format"),
			UnauthUser:    errors.New("user is not authenticated"),
			OrderFormat:   errors.New("invalid order format"),
		},
	}
}
