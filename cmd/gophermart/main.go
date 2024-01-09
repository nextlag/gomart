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
	"github.com/nextlag/gomart/internal/controller/router"
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
		log     = logger.SetupLogger()
		cfg     = config.Cfg
		er      = usecase.Status()
		r       usecase.Repository
		useCase = usecase.New(r)
		entity  = usecase.NewEntity(*useCase)
	)

	log.Debug("initialized flags",
		slog.String("-a", cfg.Host),
		slog.String("-d", cfg.DSN),
		slog.String("-k", cfg.SecretToken),
		slog.String("-l", cfg.LogLevel.String()),
		slog.String("-r", cfg.Accrual),
	)
	// Repository
	db, err := psql.New(cfg.DSN, log)
	if err != nil {
		log.Error("failed to connect in database", "error main", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	// Инициализация use case, который предоставляет бизнес-логику для обработки запросов.
	uc := usecase.New(usecase.NewStorage(db, log))

	// Создание нового маршрутизатора Chi для обработки HTTP-запросов.
	handler := chi.NewRouter()

	// Настройка маршрутов с использованием роутера и создание обработчика запросов.
	rout := router.SetupRouter(handler, log, uc, er, entity)

	// Настройка HTTP-сервера с использованием созданного маршрутизатора.
	srv := setupServer(rout)

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
