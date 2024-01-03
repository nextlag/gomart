package psql

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	usersTable = `CREATE TABLE IF NOT EXISTS users (
		login VARCHAR(255) PRIMARY KEY,
		password VARCHAR(255),
		balance FLOAT,
		withdrawn FLOAT
	);`
	ordersTable = `CREATE TABLE IF NOT EXISTS orders (
		login VARCHAR(255),
		"order" VARCHAR(255) PRIMARY KEY,
		status VARCHAR(255),
		uploaded_at TIMESTAMP,
		bonuses_withdrawn FLOAT,
		accrual FLOAT
	);`
)

type Postgres struct {
	db  *sql.DB
	log *slog.Logger
}

// CreateTable - создает таблицу в базе данных
func (s *Postgres) createTable(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, usersTable)
	if err != nil {
		return fmt.Errorf("exec create users table query: %w", err)
	}

	_, err = s.db.ExecContext(ctx, ordersTable)
	if err != nil {
		return fmt.Errorf("exec create orders table query: %w", err)
	}

	return nil
}

func New(ctx context.Context, cfg string, log *slog.Logger) (*Postgres, error) {
	// Создание подключения к базе данных с использованием контекста
	db, err := sql.Open("pgx", cfg)
	if err != nil {
		log.Error("error when opening a connection to the database", err)
		return nil, fmt.Errorf("db connection error: %w", err)
	}

	// Проверка подключения к базе данных с использованием контекста
	if err := db.PingContext(ctx); err != nil {
		log.Error("error when checking database connection", err.Error())
		return nil, fmt.Errorf("db ping error: %w", err)
	}

	storage := &Postgres{
		db:  db,
		log: log,
	}

	// Создание таблицы с использованием контекста
	if err := storage.createTable(ctx); err != nil {
		log.Error("error when creating a table in the database", err.Error())
		return nil, fmt.Errorf("create table error: %w", err)
	}

	log.Info("database connection successful", "psql.go", "func New()")

	return storage, nil
}

// Close - закрывает соединение с базой данных
func (s *Postgres) Close() {
	err := s.db.Close()
	if err != nil {
		return
	}
}
