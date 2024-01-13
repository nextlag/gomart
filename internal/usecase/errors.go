package usecase

import (
	"errors"
)

type ErrorAuth struct {
	Auth  error // Auth ошибка аутентификации
	Token error // Token неверная сигнатура токена
}

type ErrCommon struct {
	InternalServer error // InternalServer status 500: внутренняя ошибка сервера
	Request        error // Request status 400: error request
	DecodeJSON     error // DecodeJSON status 400: failed to decode json
}

type ErrAuthentication struct {
	Unauthorized error // Unauthorized status 401 incorrect login or password
	NoCookie     error // NoCookie status 500 can't set cookie
}

type ErrPostOrder struct {
	ThisUser      error // ThisUser status 200: номер заказа уже был загружен этим пользователем
	AnotherUser   error // AnotherUser status 409: номер заказа уже был загружен другим пользователем
	OrderAccepted error // OrderAccepted status 202: новый номер заказа принят в обработку
	RequestFormat error // RequestFormat status 400: неверный формат запроса
	UnAuthUser    error // UnAuthUser status 401: пользователь не аутентифицирован
	OrderFormat   error // OrderFormat status 422: неверный формат номера заказа
}

type ErrGetOrders struct {
	GetOrders error // GetOrders status 500: ошибка получения ордера
	NoContent error // NoContent status 204: нет данных для ответа
}
type ErrDebit struct {
	NoBalance error // NoBalance недостаточно баланса
}

type AllErr struct {
	*ErrorAuth
	*ErrCommon
	*ErrAuthentication
	*ErrPostOrder
	*ErrGetOrders
	*ErrDebit
}

func NewErr() *AllErr {
	return &AllErr{
		&ErrorAuth{
			Auth:  errors.New("authentication error"),
			Token: errors.New("signature is invalid"),
		},
		&ErrCommon{
			InternalServer: errors.New("insert server error"),
			Request:        errors.New("error request"),
			DecodeJSON:     errors.New("failed to decode json"),
		},
		&ErrAuthentication{
			Unauthorized: errors.New("incorrect login or password"),
			NoCookie:     errors.New("can't set cookie"),
		},
		&ErrPostOrder{
			ThisUser:      errors.New("the order number has already been uploaded by this user"),
			AnotherUser:   errors.New("the order number has already been uploaded by another user"),
			OrderAccepted: errors.New("new order number accepted for processing"),
			RequestFormat: errors.New("invalid request format"),
			UnAuthUser:    errors.New("user is not authenticated"),
			OrderFormat:   errors.New("invalid order format"),
		},
		&ErrGetOrders{
			GetOrders: errors.New("error getting orders"),
			NoContent: errors.New("no information to answer"),
		},
		&ErrDebit{
			NoBalance: errors.New("not enough balance"),
		},
	}
}
