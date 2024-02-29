// Package entity represents the main business logic structures
package entity

import "time"

// User структура, предназначенная для вставки данных в таблицу пользователей
type User struct {
	Login     string  `json:"login"`
	Password  string  `json:"password"`
	Balance   float32 `json:"balance"`
	Withdrawn float32 `json:"withdrawn"`
}

// Order структура, предназначенная для вставки данных в таблицу заказов.
type Order struct {
	UserName         string    `json:"user_name,omitempty"`
	Order            string    `json:"number"`
	Status           string    `json:"status"`
	Accrual          float32   `json:"accrual,omitempty"`
	UploadedAt       time.Time `json:"uploaded_at"`
	BonusesWithdrawn float32   `json:"bonuses_withdrawn,omitempty"`
}

type AllEntity struct {
	*User
	*Order
}
