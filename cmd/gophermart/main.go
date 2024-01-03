package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/nextlag/gomart/internal/config"
	"github.com/nextlag/gomart/internal/controller/router"
	"github.com/nextlag/gomart/internal/mw/gzip"
	"github.com/nextlag/gomart/internal/mw/logger"
	"github.com/nextlag/gomart/internal/repository/psql"
	"github.com/nextlag/gomart/internal/usecase"
)

const createTablesTimeout = time.Second * 5

func setupServer(router http.Handler) *http.Server {
	// Создание HTTP-сервера с указанным адресом и обработчиком маршрутов
	return &http.Server{
		Addr:    config.Cfg.Host, // Получение адреса из настроек
		Handler: router,
	}
}

func main() {
	if err := config.MakeConfig(); err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), createTablesTimeout)
	defer cancel()

	var (
		log = logger.SetupLogger()
		cfg = config.Cfg
	)
	// TODO: drop
	// var r       usecase.Repository
	// var useCase = usecase.New(r)
	// var entity  = usecase.NewEntity(*useCase)
	// entity.Time = time.Now().Format("15:04:05 02.01.2006")
	// log.Debug("checking the transfer of data into the Entity structure", "current time", entity.Time)

	log.Debug("initialized flags",
		slog.String("-a", cfg.Host),
		slog.String("-d", cfg.DSN),
		slog.String("-k", cfg.SecretToken),
		slog.String("-r", cfg.Accrual),
		slog.String("-l", cfg.LogLevel.String()),
	)
	// Repository
	db, err := psql.New(ctx, cfg.DSN, log)
	if err != nil {
		log.Error("failed to connect in database", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	// Создание нового маршрутизатора Chi для обработки HTTP-запросов.
	handler := chi.NewRouter()

	// Инициализация use case, который предоставляет бизнес-логику для обработки запросов.
	uc := usecase.New(usecase.NewStorage(db))

	// Настройка маршрутов с использованием роутера и создание обработчика запросов.
	rout := router.SetupRouter(handler, log, uc)

	// Создание обработчика для поддержки сжатия gzip при обработке HTTP-запросов.
	mv := gzip.New(rout.ServeHTTP)

	// Настройка HTTP-сервера с использованием созданного маршрутизатора.
	srv := setupServer(mv)

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
