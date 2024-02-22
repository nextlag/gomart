// Package controllers provides HTTP request handlers for various endpoints in the application.
// These handlers are responsible for processing incoming requests, invoking corresponding
// methods from the use case layer, and returning appropriate responses.
package controllers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/nextlag/gomart/internal/mw/auth"
	"github.com/nextlag/gomart/internal/mw/gzip"
	"github.com/nextlag/gomart/internal/mw/logger"
	"github.com/nextlag/gomart/internal/usecase"
)

//go:generate mockgen -destination=mocks/mocks.go -package=mocks github.com/nextlag/gomart/internal/controllers UseCase
type UseCase interface {
	Do() *usecase.UseCase
	DoRegister(ctx context.Context, login, password string, r *http.Request) error
	DoAuth(ctx context.Context, login, password string, r *http.Request) error
	DoInsertOrder(ctx context.Context, user, order string) error
	DoGetOrders(ctx context.Context, user string) ([]byte, error)
	DoGetBalance(ctx context.Context, login string) (float32, float32, error)
	DoDebit(ctx context.Context, user, numOrder string, sum float32) error
	DoGetWithdrawals(ctx context.Context, user string) ([]byte, error)
}

type Controller struct {
	uc  UseCase
	log usecase.Logger
}

func New(uc UseCase, log usecase.Logger) *Controller {
	return &Controller{uc: uc, log: log}
}

// Router настраивает маршруты для обработчика запросов.
//
// Этот метод настраивает маршруты для обработки HTTP-запросов с использованием роутера chi.
// Он устанавливает несколько middleware для обработки запросов, включая логирование, сжатие и восстановление
// после паники. Затем он определяет несколько конечных точек API.
//
// Параметры:
//   - handler: *chi.Mux - роутер, к которому добавляются маршруты.
//
// Возвращаемые значения:
//   - *chi.Mux - роутер с настроенными маршрутами.
func (c *Controller) Router(handler *chi.Mux) *chi.Mux {
	// Использование middleware для обработки запросов
	handler.Use(middleware.RequestID)
	handler.Use(logger.New(c.log))
	handler.Use(middleware.Logger)
	handler.Use(gzip.New())
	handler.Use(middleware.Recoverer)

	// Группировка маршрутов по аутентификации пользователя
	handler.Group(func(r chi.Router) {
		// Регистрация и аутентификация пользователя
		r.Post("/api/user/register", c.Register)
		r.Post("/api/user/login", c.Authentication)

		// Группа маршрутов, требующих аутентификации пользователя
		r.With(auth.CookieAuthentication(c.log, c.uc.Do().Err())).Group(func(r chi.Router) {
			// Маршруты для работы с заказами, балансом и выводом средств
			r.Post("/api/user/orders", c.PostOrders)
			r.Post("/api/user/balance/withdraw", c.Withdraw)
			r.Get("/api/user/withdrawals", c.Withdrawals)
			r.Get("/api/user/balance", c.Balance)
			r.Get("/api/user/orders", c.GetOrders)
		})
	})

	return handler
}
