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
		log = slogpretty.SetupLogger(config.ProjectRoot)
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
	uc := usecase.New(db, cfg)

	// init controllers
	controller := controllers.New(uc, log)
	r := chi.NewRouter()
	r.Mount("/", controller.Router(r))

	// init server
	srv := setupServer(r)
	log.Info("server starting", slog.String("host", srv.Addr))

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	stop := make(chan struct{})
	defer close(stop)

	// go func() {
	// 	if err := db.Sync(stop); err != nil {
	// 		log.Error("db.Sync()", "error", err.Error())
	// 		sigs <- os.Interrupt
	// 		return
	// 	}
	// }()

	go func() {
		// Закрытие канала stop при завершении работы функции
		defer close(stop)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("failed to start server", "error", err.Error())
			sigs <- os.Interrupt
			return
		}
	}()

	log.Info("server started")
	<-sigs
	log.Info("server stopped")
}
