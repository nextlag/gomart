package usecase

import "errors"

var (
	NoLogin        = errors.New("login is already taken")
	Auth           = errors.New("authentication error")
	Token          = errors.New("signature is invalid")
	InternalServer = errors.New("insert server error")
	Request        = errors.New("error request")
	DecodeJSON     = errors.New("failed to decode json")
	Unauthorized   = errors.New("incorrect login or password")
	NoCookie       = errors.New("can't set cookie")
	OrderNotFound  = errors.New("no such order exists")
	ThisUser       = errors.New("the order number has already been uploaded by this user")
	AnotherUser    = errors.New("the order number has already been uploaded by another user")
	OrderAccepted  = errors.New("new order number accepted for processing")
	RequestFormat  = errors.New("invalid request format")
	UnAuthUser     = errors.New("user is not authenticated")
	OrderFormat    = errors.New("invalid order format")
	GetOrders      = errors.New("error getting orders")
	NoContent      = errors.New("no information to answer")
	NoBalance      = errors.New("not enough balance")
)

type UCErr struct {
	NoLogin        error
	Auth           error
	Token          error
	InternalServer error
	Request        error
	DecodeJSON     error
	Unauthorized   error
	NoCookie       error
	OrderNotFound  error
	ThisUser       error
	AnotherUser    error
	OrderAccepted  error
	RequestFormat  error
	UnAuthUser     error
	OrderFormat    error
	GetOrders      error
	NoContent      error
	NoBalance      error
}

func (uc *UseCase) Er() *UCErr {
	return &UCErr{
		NoLogin:        NoLogin,
		Auth:           Auth,
		Token:          Token,
		InternalServer: InternalServer,
		Request:        Request,
		DecodeJSON:     DecodeJSON,
		Unauthorized:   Unauthorized,
		NoCookie:       NoCookie,
		OrderNotFound:  OrderNotFound,
		ThisUser:       ThisUser,
		AnotherUser:    AnotherUser,
		OrderAccepted:  OrderAccepted,
		RequestFormat:  RequestFormat,
		UnAuthUser:     UnAuthUser,
		OrderFormat:    OrderFormat,
		GetOrders:      GetOrders,
		NoContent:      NoContent,
		NoBalance:      NoBalance,
	}
}
