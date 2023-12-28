package main

import (
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nextlag/gomart/internal/config"
	"github.com/nextlag/gomart/internal/controller"
	"github.com/nextlag/gomart/internal/controller/router"
	"github.com/nextlag/gomart/internal/mw/gzip"
	"github.com/nextlag/gomart/internal/mw/logger"
	"github.com/nextlag/gomart/internal/usecase"
)

func setupServer(router http.Handler) *http.Server {
	// Создание HTTP-сервера с указанным адресом и обработчиком маршрутов
	return &http.Server{
		Addr:    config.Cfg.Host, // Получение адреса из настроек
		Handler: router,
	}
}

func main() {
	if err := config.MakeConfig(); err != nil {
		log.Fatal(err)
	}

	var (
		uc      controller.UseCase
		r       usecase.Repository
		useCase = usecase.New(r)
		entity  = usecase.NewEntity(*useCase)
		log     = logger.SetupLogger()
		cfg     = config.Cfg
		rout    = router.SetupRouter(uc, log)
		mv      = gzip.New(rout.ServeHTTP)
		srv     = setupServer(mv)
	)
	// TODO: drop
	entity.Time = time.Now().Format("15:04:05 02.01.2006")
	log.Debug("checking the transfer of data into the Entity structure", "current time", entity.Time)

	log.Debug("initialized flags",
		slog.String("-a", cfg.Host),
		slog.String("-d", cfg.DSN),
		slog.String("-k", cfg.SecretToken),
		slog.String("-r", cfg.Accrual),
		slog.String("-l", cfg.LogLevel.String()),
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
