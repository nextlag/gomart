package entity

import "time"

// User - отражает информацию о зарегистрированных пользователях
type User struct {
	Login     string  `json:"login"`
	Password  string  `json:"password"`
	Balance   float32 `json:"balance"`
	Withdrawn float32 `json:"withdrawn"`
}

// Orders - cтруктура, предназначенная для вставки данных в таблицу заказов.
type Orders struct {
	Users            string    `json:"users,omitempty"`
	Order            string    `json:"number"`
	Status           string    `json:"status"`
	Accrual          float32   `json:"accrual,omitempty"`
	UploadedAt       time.Time `json:"uploaded_at"`
	BonusesWithdrawn float32   `json:"bonuses_withdrawn,omitempty"`
}

type AllEntity struct {
	*User
	*Orders
}
