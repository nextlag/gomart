// Package usecase containing major errors
package usecase

import (
	"errors"
)

type ErrAll struct {
	ErrNoLogin        error
	ErrAuth           error
	ErrToken          error
	ErrInternalServer error
	ErrRequest        error
	ErrDecodeJSON     error
	ErrUnauthorized   error
	ErrNoCookie       error
	ErrOrderNotFound  error
	ErrThisUser       error
	ErrAnotherUser    error
	ErrOrderAccepted  error
	ErrRequestFormat  error
	ErrUnAuthUser     error
	ErrOrderFormat    error
	ErrGetOrders      error
	ErrNoContent      error
	ErrNoBalance      error
	ErrNoRows         error
}

var (
	ErrNoLogin        = errors.New("login is already taken")
	ErrAuth           = errors.New("authentication error")
	ErrToken          = errors.New("signature is invalid")
	ErrInternalServer = errors.New("insert server error")
	ErrRequest        = errors.New("error request")
	ErrDecodeJSON     = errors.New("failed to decode json")
	ErrUnauthorized   = errors.New("incorrect login or password")
	ErrNoCookie       = errors.New("can't set cookie")
	ErrOrderNotFound  = errors.New("no such order exists")
	ErrThisUser       = errors.New("the order number has already been uploaded by this user")
	ErrAnotherUser    = errors.New("the order number has already been uploaded by another user")
	ErrOrderAccepted  = errors.New("new order number accepted for processing")
	ErrRequestFormat  = errors.New("invalid request format")
	ErrUnAuthUser     = errors.New("user is not authenticated")
	ErrOrderFormat    = errors.New("invalid order format")
	ErrGetOrders      = errors.New("error getting orders")
	ErrNoContent      = errors.New("no information to answer")
	ErrNoBalance      = errors.New("not enough balance")
	ErrNoRows         = errors.New("no rows were found")
)

func (uc *UseCase) Err() *ErrAll {
	return &ErrAll{
		ErrNoLogin:        ErrNoLogin,
		ErrAuth:           ErrAuth,
		ErrToken:          ErrToken,
		ErrInternalServer: ErrInternalServer,
		ErrRequest:        ErrRequest,
		ErrDecodeJSON:     ErrDecodeJSON,
		ErrUnauthorized:   ErrUnauthorized,
		ErrNoCookie:       ErrNoCookie,
		ErrOrderNotFound:  ErrOrderNotFound,
		ErrThisUser:       ErrThisUser,
		ErrAnotherUser:    ErrAnotherUser,
		ErrOrderAccepted:  ErrOrderAccepted,
		ErrRequestFormat:  ErrRequestFormat,
		ErrUnAuthUser:     ErrUnAuthUser,
		ErrOrderFormat:    ErrOrderFormat,
		ErrGetOrders:      ErrGetOrders,
		ErrNoContent:      ErrNoContent,
		ErrNoBalance:      ErrNoBalance,
		ErrNoRows:         ErrNoRows,
	}
}
