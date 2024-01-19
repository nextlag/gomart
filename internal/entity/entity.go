package entity

import "time"

// User - отражает информацию о зарегистрированных пользователях
type User struct {
	Login     string  `json:"login"`
	Password  string  `json:"password"`
	Balance   float64 `json:"balance"`
	Withdrawn float64 `json:"withdrawn"`
}

// Orders - cтруктура, предназначенная для вставки данных в таблицу заказов.
type Orders struct {
	Users            string    `bun:"users" json:"users,omitempty"`
	Number           string    `bun:"number" json:"number"`
	Status           string    `bun:"status" json:"status"`
	Accrual          float64   `json:"accrual,omitempty"`
	UploadedAt       time.Time `bun:"uploaded_at" json:"uploaded_at"`
	BonusesWithdrawn float64   `bun:"bonuses_withdrawn" json:"bonuses_withdrawn,omitempty"`
}

type AllEntity struct {
	*User
	*Orders
}

// Withdrawals - cтруктура, предназначенная для возврата клиенту данных о заказах с снятыми бонусами.
type Withdrawals struct {
	Number           string    `json:"number"`
	Time             time.Time `json:"processed_at"`
	BonusesWithdrawn float64   `json:"sum"`
}

// cтруктура, предназначенная для получения данных из системы начисления
type OrderUpdateFromAccrual struct {
	Number  string  `json:"number"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}
