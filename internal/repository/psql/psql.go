package psql

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	createTablesTimeout = time.Second * 5
	usersTable          = `CREATE TABLE IF NOT EXISTS users (
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
	DB *sql.DB
}

// CreateTable - создает таблицу в базе данных
func (s *Postgres) createTable(ctx context.Context) error {
	_, err := s.DB.ExecContext(ctx, usersTable)
	if err != nil {
		return fmt.Errorf("exec create users table query: %w", err)
	}

	_, err = s.DB.ExecContext(ctx, ordersTable)
	if err != nil {
		return fmt.Errorf("exec create orders table query: %w", err)
	}

	return nil
}

func New(cfg string, log *slog.Logger) (*Postgres, error) {
	ctx, cancel := context.WithTimeout(context.Background(), createTablesTimeout)
	defer cancel()
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
		DB: db,
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
	err := s.DB.Close()
	if err != nil {
		return
	}
}
