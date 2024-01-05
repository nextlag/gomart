package router

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/nextlag/gomart/internal/controller"
	"github.com/nextlag/gomart/internal/mw/auth"
	"github.com/nextlag/gomart/internal/mw/logger"
	"github.com/nextlag/gomart/internal/usecase"
)

func SetupRouter(handler *chi.Mux, log *slog.Logger, useCase *usecase.UseCase) *chi.Mux {
	handler.Use(middleware.RequestID)
	handler.Use(logger.New(log))

	h := controller.New(useCase, log)

	// Общая группа для роутов, где требуется аутентификация
	handler.Group(func(r chi.Router) {
		r.Post("/api/user/register", h.Register)
		r.Post("/api/user/login", h.Login)

		// Группа с применением WithCookieLogin ко всем роутам внутри
		r.With(auth.WithCookieLogin(log)).Group(func(r chi.Router) {
			r.Post("/api/user/orders", h.PostOrders)
			r.Post("/api/user/balance/withdraw", h.Withdraw)
			r.Get("/api/user/balance", h.Balance)
			r.Get("/api/user/orders", h.GetOrders)
		})
	})

	// Роут без аутентификации
	handler.Get("/api/user/withdrawals", h.Withdrawals)

	return handler
}
