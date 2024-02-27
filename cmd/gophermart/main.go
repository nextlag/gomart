package main

import (
	"context"
	"errors"
	stdLog "log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-chi/chi/v5"

	"github.com/nextlag/gomart/internal/config"
	"github.com/nextlag/gomart/internal/controllers"
	"github.com/nextlag/gomart/internal/repository/psql"
	"github.com/nextlag/gomart/internal/usecase"
	"github.com/nextlag/gomart/pkg/logger/slogpretty"
)

func setupServer(router http.Handler) *http.Server {
	return &http.Server{
		Addr:    config.Cfg.Host,
		Handler: router,
	}
}

func main() {
	if err := config.MakeConfig(); err != nil {
		stdLog.Fatal(err)
	}

	var (
		log = slogpretty.SetupLogger(config.Cfg.ProjectRoot)
		cfg = config.Cfg
	)

	log.Debug("initialized flags",
		slog.String("-a", cfg.Host),
		slog.String("-d", cfg.DSN),
		slog.String("-k", cfg.SecretToken),
		slog.String("-l", cfg.LogLevel.String()),
		slog.String("-r", cfg.Accrual),
		slog.String("-p", cfg.ProjectRoot),
	)

	// init repository
	db, err := psql.New(cfg.DSN, log)
	if err != nil {
		log.Error("failed to connect in database", "error main", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	// init usecase
	uc := usecase.New(db, cfg)

	// init controllers
	controller := controllers.New(uc, log)
	r := chi.NewRouter()
	r.Mount("/", controller.Router(r))

	// init server
	srv := setupServer(r)
	log.Info("server starting", slog.String("host", srv.Addr))

	// WaitGroup для ожидания завершения работы горутин
	var wg sync.WaitGroup
	wg.Add(2)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	stop := make(chan struct{})

	go func() {
		defer wg.Done()
		if err = db.Sync(stop); err != nil {
			log.Error("db.Sync()", "error", err.Error())
			sigs <- os.Interrupt
			return
		}
	}()

	go func() {
		defer wg.Done()
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("failed to start server", "error", err.Error())
			sigs <- os.Interrupt
			return
		}
	}()

	// Ожидание получения сигнала от OS
	<-sigs

	// Закрытие канала stop и остановка http-сервера
	close(stop)
	// TODO
	ctx := context.TODO()
	if err = srv.Shutdown(ctx); err != nil {
		log.Error("server shutdown error", "error", err.Error())
	}

	// Ожидание завершения работы горутин
	wg.Wait()

	log.Info("server stopped")
}
