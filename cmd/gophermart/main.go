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
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/nextlag/gomart/internal/config"
	"github.com/nextlag/gomart/internal/controllers"
	"github.com/nextlag/gomart/internal/repository/psql"
	"github.com/nextlag/gomart/internal/usecase"
	"github.com/nextlag/gomart/pkg/logger/l"
)

func main() {
	if err := config.MakeConfig(); err != nil {
		stdLog.Fatal(err)
	}
	ctx, cansel := context.WithCancel(context.Background())
	// defer cansel()
	ctx = l.ContextWithLogger(ctx, l.LoggerNew(config.Cfg.ProjectRoot))

	var (
		log = l.L(ctx)
		cfg = config.Cfg
	)

	log.Debug("initialized flags",
		l.StringAttr("-a", cfg.Host),
		l.StringAttr("-d", cfg.DSN),
		l.StringAttr("-k", cfg.SecretToken),
		l.StringAttr("-l", cfg.LogLevel.String()),
		l.StringAttr("-r", cfg.Accrual),
		l.StringAttr("-p", cfg.ProjectRoot),
	)

	// init repository
	db, err := psql.New(ctx, cfg.DSN)
	if err != nil {
		log.Error("failed to connect in database", l.ErrAttr(err))
		os.Exit(1)
	}
	defer db.Close()

	// init usecase
	uc := usecase.New(db, cfg)

	r := chi.NewRouter()

	// create new server
	srv := controllers.New(ctx, uc).NewServer(r)
	r.Mount("/", srv.Handler)

	log.Info("server starting", slog.String("host", srv.Addr))

	// WaitGroup для ожидания завершения работы горутин
	var wg sync.WaitGroup
	wg.Add(2)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer wg.Done()
		if err = db.Sync(ctx); err != nil {
			log.Error("db.Sync()", l.ErrAttr(err))
			sigs <- os.Interrupt
			return
		}
	}()

	go func() {
		defer wg.Done()
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("failed to start server", l.ErrAttr(err))
			sigs <- os.Interrupt
			return
		}
	}()

	// Ожидание получения сигнала от OS
	<-sigs

	// Закрытие канала stop и остановка http-сервера
	cansel()
	ctxTime, canselShutdown := context.WithTimeout(context.Background(), time.Second*10)
	defer canselShutdown()
	if err = srv.Shutdown(ctxTime); err != nil {
		log.Error("server shutdown error", l.ErrAttr(err))
	}

	// Ожидание завершения работы горутин
	wg.Wait()

	log.Info("server stopped")
}
