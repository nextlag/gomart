package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/nextlag/gomart/internal/entity"
)

var (
	// Ошибка, указывающая на то, что заказ загружен другим пользователем.
	ErrAlreadyLoadedOrder = errors.New("the order number has already been uploaded by another user")
	// Ошибка, указывающая на то, что заказ загружен пользователем.
	ErrYouAlreadyLoadedOrder = errors.New("the order number has already been uploaded by this user")
	// Ошибка, указывающая на то, что у пользователя недостаточно баланса.
	ErrNotEnoughBalance = errors.New("not enough balance")
	// Ошибка, указывающая, что строки не найдены.
	ErrNoRows = errors.New("no rows were found")
)

type Repository interface {
	// Register - регистрация нового пользователя
	Register(ctx context.Context, login, password string) error
	// Auth — проверяет, есть ли совпадение в базе по логину и паролю.
	Auth(ctx context.Context, login, password string) error
	// InsertOrder - используется для вставки информации о заказе в базу данных.
	InsertOrder(ctx context.Context, login string, order string) error
}

type UseCase struct {
	r Repository // interface Repository
	e *entity.Entity
}

func New(r Repository) *UseCase {
	e := &entity.Entity{}
	return &UseCase{r, e}
}

// func NewEntity(uc UseCase) *entity.Entity {
// 	return uc.e
// }

func (uc *UseCase) DoRegister(ctx context.Context, login, password string, r *http.Request) error {
	err := uc.r.Register(ctx, login, password)
	if err != nil {
		return fmt.Errorf("failed to push registration data %v", err.Error())
	}
	return nil
}
func (uc *UseCase) DoAuth(ctx context.Context, login, password string, r *http.Request) error {
	err := uc.r.Auth(ctx, login, password)
	if err != nil {
		return fmt.Errorf("failed to push registration data %v", err.Error())
	}
	return nil
}
func (uc *UseCase) DoInsertOrder(ctx context.Context, login string, order string) error {
	err := uc.r.InsertOrder(ctx, login, order)
	if err != nil {
		return fmt.Errorf("failed to push order number %v", err.Error())
	}
	return nil
}
