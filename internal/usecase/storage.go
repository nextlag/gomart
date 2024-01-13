package usecase

import (
	"context"
	"fmt"
)

const (
	usersTable = `CREATE TABLE IF NOT EXISTS users (
		login VARCHAR(255) PRIMARY KEY,
		password VARCHAR(255),
		balance FLOAT,
		withdrawn FLOAT
	);`
	ordersTable = `CREATE TABLE IF NOT EXISTS orders (
		users VARCHAR(255),
		"order" VARCHAR(255) PRIMARY KEY,
		status VARCHAR(255),
		accrual FLOAT,
		uploaded_at TIMESTAMP,
		bonuses_withdrawn FLOAT
	);`
)

// CreateTable - создает таблицу в базе данных
func (uc *UseCase) CreateTable(ctx context.Context) error {
	_, err := uc.DB.ExecContext(ctx, usersTable)
	if err != nil {
		return fmt.Errorf("exec create users table query: %v", err.Error())
	}

	_, err = uc.DB.ExecContext(ctx, ordersTable)
	if err != nil {
		return fmt.Errorf("exec create orders table query: %v", err.Error())
	}

	return nil
}

// Close - закрывает соединение с базой данных
func (uc *UseCase) Close() {
	err := uc.DB.Close()
	if err != nil {
		return
	}
}
