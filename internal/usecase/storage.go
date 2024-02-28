// Package usecase - creating database tables
package usecase

import (
	"context"
	"fmt"
)

const (
	usersTable = `CREATE TABLE IF NOT EXISTS users (
		login VARCHAR(255) PRIMARY KEY,
		password VARCHAR(255),
		balance FLOAT not null ,
		withdrawn FLOAT not null 
	);`
	ordersTable = `CREATE TABLE IF NOT EXISTS orders (
		"user_name" VARCHAR(255),
		"order" VARCHAR(255) PRIMARY KEY,
		status VARCHAR(255),
		accrual FLOAT not null ,
		uploaded_at TIMESTAMP,
		bonuses_withdrawn FLOAT
	);`
)

// CreateTable - creating tables in the database
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

// Close - closing the connection to the database
func (uc *UseCase) Close() {
	err := uc.DB.Close()
	if err != nil {
		return
	}
}
