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

	// init server
	srv := setupServer(r)

	log.Info("server starting", slog.String("host", srv.Addr))

	// Создание канала для получения сигналов операционной системы
	sigs := make(chan os.Signal, 1)
	// Уведомление канала о сигналах прерывания (Ctrl+C) и завершения работы
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Создание канала для остановки операций
	stop := make(chan struct{})
	// Отложенное закрытие канала stop при выходе из функции main
	defer close(stop)

	// Запуск функции db.Sync() в горутине для синхронизации с базой данных
	go func() {
		// Выполнение синхронизации с базой данных с передачей канала остановки
		if err := db.Sync(stop); err != nil {
			log.Error("error in db.Sync()", "error", err.Error())
		}
	}()

	// Запуск HTTP-сервера в горутине
	go func() {
		// Закрытие канала stop при завершении работы функции
		defer close(stop)
		// Запуск HTTP-сервера и обработка ошибок
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// В случае ошибки при запуске сервера, запись ошибки в лог и отправка сигнала прерывания
			log.Error("failed to start server", "error main", err.Error())
			sigs <- os.Interrupt
			return
		}
	}()

	log.Info("server started")
	// Ожидание получения сигнала завершения работы сервера
	<-sigs
	log.Info("server stopped")

}
