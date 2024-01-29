// Package psql DB initialization
package psql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/nextlag/gomart/internal/usecase"
	"github.com/nextlag/gomart/pkg/logger/l"
)

const createTablesTimeout = time.Second * 5

// New - postgreSQL DB initialization
func New(ctx context.Context, cfg string) (*usecase.UseCase, error) {
	log := l.L(ctx)
	ctx, cancel := context.WithTimeout(context.Background(), createTablesTimeout)
	defer cancel()

	db, err := sql.Open("postgres", cfg)
	if err != nil {
		log.Error("error when opening a connection to the database", l.ErrAttr(err))
		return nil, fmt.Errorf("db connection error: %v", err.Error())
	}

	if err = db.PingContext(ctx); err != nil {
		log.Error("error when checking database connection", l.ErrAttr(err))
		return nil, fmt.Errorf("db ping error: %v", err.Error())
	}

	storage := &usecase.UseCase{
		DB: db,
	}

	if err = storage.CreateTable(ctx); err != nil {
		log.Error("error when creating a table in the database", l.ErrAttr(err))
		return nil, fmt.Errorf("create table error: %v", err.Error())
	}

	log.Info("db connection success")

	return storage, nil
}
