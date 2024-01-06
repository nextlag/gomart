package usecase

import "errors"

var (
	// ErrThisUser - status 200: номер заказа уже был загружен этим пользователем
	ErrThisUser = errors.New("the order number has already been uploaded by this user")
	// ErrOrderAccepted - status 202: новый номер заказа принят в обработку
	ErrOrderAccepted = errors.New("new order number accepted for processing")
	// ErrRequestFormat - status 400: неверный формат запроса
	ErrRequestFormat = errors.New("invalid request format")
	// ErrUnauthUser - status 401: пользователь не аутентифицирован
	ErrUnauthUser = errors.New("user is not authenticated")
	// ErrAnotherUser - status 409: номер заказа уже был загружен другим пользователем
	ErrAnotherUser = errors.New("the order number has already been uploaded by another user")
	// ErrOrderFormat - status 422: неверный формат номера заказа
	ErrOrderFormat = errors.New("invalid order format")
	// ErrInternalServer - status 500: внутренняя ошибка сервера
	ErrInternalServer = errors.New("insert server error")
	// ErrNoBalance - недостаточно баланса.
	ErrNoBalance = errors.New("not enough balance")
	// ErrNoRows - строки не найдены
	ErrNoRows = errors.New("no rows were found")
)
