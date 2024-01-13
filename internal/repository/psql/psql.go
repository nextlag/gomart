package psql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/nextlag/gomart/internal/usecase"
)

const createTablesTimeout = time.Second * 5

func New(cfg string, log usecase.Logger) (*usecase.UseCase, error) {
	ctx, cancel := context.WithTimeout(context.Background(), createTablesTimeout)
	defer cancel()
	// Создание подключения к базе данных с использованием контекста
	db, err := sql.Open("postgres", cfg)
	if err != nil {
		log.Error("error when opening a connection to the database", "error psql", err.Error())
		return nil, fmt.Errorf("db connection error: %v", err.Error())
	}

	// Проверка подключения к базе данных с использованием контекста
	if err := db.PingContext(ctx); err != nil {
		log.Error("error when checking database connection", "error psql", err.Error())
		return nil, fmt.Errorf("db ping error: %v", err.Error())
	}

	storage := &usecase.UseCase{
		DB: db,
	}

	// Создание таблицы с использованием контекста
	if err := storage.CreateTable(ctx); err != nil {
		log.Error("error when creating a table in the database", "error psql", err.Error())
		return nil, fmt.Errorf("create table error: %v", err.Error())
	}

	log.Info("db connection success")

	return storage, nil
}
