package router

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/nextlag/gomart/internal/controller"
	"github.com/nextlag/gomart/internal/mw/auth"
	"github.com/nextlag/gomart/internal/mw/gzip"
	"github.com/nextlag/gomart/internal/mw/logger"
	"github.com/nextlag/gomart/internal/usecase"
)

func SetupRouter(handler *chi.Mux, log *slog.Logger, uc *usecase.UseCase) *chi.Mux {
	handler.Use(middleware.RequestID)
	handler.Use(logger.New(log))
	handler.Use(middleware.Logger)
	handler.Use(gzip.New())

	h := controller.New(uc, log)

	handler.Group(func(r chi.Router) {
		r.Post("/api/user/register", h.Register)
		r.Post("/api/user/login", h.Login)
		r.With(auth.CookieAuthentication(uc, log)).Group(func(r chi.Router) {
			r.Post("/api/user/orders", h.PostOrders)
			r.Post("/api/user/balance/withdraw", h.Withdraw)
			r.Post("/api/user/withdrawals", h.Withdrawals)
			r.Get("/api/user/balance", h.Balance)
			r.Get("/api/user/orders", h.GetOrders)
		})
	})
	return handler
}
