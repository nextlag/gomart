package main

import (
	"errors"
	stdLog "log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"

	"github.com/nextlag/gomart/internal/config"
	"github.com/nextlag/gomart/internal/controllers"
	"github.com/nextlag/gomart/internal/mw/logger"
	"github.com/nextlag/gomart/internal/repository/psql"
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
		stdLog.Fatal(err)
	}

	var (
		log = logger.SetupLogger()
		cfg = config.Cfg
	)

	log.Debug("initialized flags",
		slog.String("-a", cfg.Host),
		slog.String("-d", cfg.DSN),
		slog.String("-k", cfg.SecretToken),
		slog.String("-l", cfg.LogLevel.String()),
		slog.String("-r", cfg.Accrual),
	)

	// init repository
	db, err := psql.New(cfg.DSN, log)
	if err != nil {
		log.Error("failed to connect in database", "error main", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	// init usecase
	uc := usecase.New(db, log, cfg)

	// init controllers
	controller := controllers.New(uc, log)
	r := chi.NewRouter()
	r.Mount("/", controller.Router(r))

	go func() {
		if err = db.Sync(); err != nil {
			// Обработка ошибок, например, логирование
			log.Error("error in uc.Sync()", "error", err.Error())
		}
	}()

	// init server
	srv := setupServer(r)

	log.Info("server starting", slog.String("host", srv.Addr))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Запуск HTTP-сервера в горутине
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			// Если сервер не стартовал, логируем ошибку
			log.Error("failed to start server", "error main", err.Error())
			done <- os.Interrupt
		}
	}()
	log.Info("server started")

	<-done // Ожидание сигнала завершения
	log.Info("server stopped")
}
