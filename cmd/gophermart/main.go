package main

import (
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/nextlag/gomart/internal/config"
	"github.com/nextlag/gomart/internal/handlers"
	"github.com/nextlag/gomart/internal/handlers/router"
	"github.com/nextlag/gomart/internal/mw/logger"
)

func main() {
	if err := config.MakeConfig(); err != nil {
		log.Fatal(err)
	}

	var (
		uc   handlers.UseCase
		log  = logger.SetupLogger()
		cfg  = config.Cfg
		rout = router.SetupRouter(uc, log)
	)

	srv := http.Server{
		Addr:    cfg.Host,
		Handler: rout,
	}

	log.Debug("initialized flags",
		slog.String("-a", cfg.Host),
		slog.String("-d", cfg.DSN),
		slog.String("-l", cfg.LogLevel.String()),
		slog.String("-r", cfg.Accrual),
	)

	log.Info("server starting", slog.String("host", srv.Addr))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Запуск HTTP-сервера в горутине
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			// Если сервер не стартовал, логируем ошибку
			log.Error("failed to start server", slog.String("error", err.Error()))
			done <- os.Interrupt
		}
	}()
	log.Info("server started")

	<-done // Ожидание сигнала завершения
	log.Info("server stopped")
}
