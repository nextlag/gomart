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

	// Используем контроллер в качестве хендлера
	handler.Route("/", func(r chi.Router) {
		r.Post("/api/user/register", h.Register)
		r.Post("/api/user/login", h.Login)
		r.With(auth.WithCookieLogin(log)).Post("/api/user/orders", h.PostOrders)
		r.With(auth.WithCookieLogin(log)).Post("/api/user/balance/withdraw", h.Withdraw)

		r.With(auth.WithCookieLogin(log)).Get("/api/user/balance", h.Balance)
		r.With(auth.WithCookieLogin(log)).Get("/api/user/orders", h.GetOrders)
		r.Get("/api/user/withdrawals", h.Withdrawals)
	})
	return handler
}
