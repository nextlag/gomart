package router

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/nextlag/gomart/internal/handlers"
	"github.com/nextlag/gomart/internal/mw/logger"
)

func SetupRouter(uc handlers.UseCase, log *slog.Logger) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)

	h := handlers.New(uc)

	// Используем контроллер в качестве хендлера
	router.With(logger.New(log)).Route("/", func(r chi.Router) {
		r.Get("/api/user/balance", h.Balance)
		r.Get("/api/user/orders", h.GetOrders)
		r.Post("/api/user/login", h.Login)
		r.Post("/api/user/orders", h.PostOrders)
		r.Post("/api/user/register", h.Register)
		r.Post("/api/user/register", h.Register)
		r.Post("/api/user/balance/withdraw", h.Withdraw)
		r.Get("/api/user/withdrawals", h.Withdrawals)
	})
	return router
}
