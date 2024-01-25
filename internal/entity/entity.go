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
	Users            string    `bun:"users" json:"users,omitempty"`
	Order            string    `bun:"order" json:"number"`
	Status           string    `bun:"status" json:"status"`
	Accrual          float32   `json:"accrual,omitempty"`
	UploadedAt       time.Time `bun:"uploaded_at" json:"uploaded_at"`
	BonusesWithdrawn float32   `bun:"bonuses_withdrawn" json:"bonuses_withdrawn,omitempty"`
}

type AllEntity struct {
	*User
	*Orders
}

// Withdrawals - cтруктура, предназначенная для возврата клиенту данных о заказах со снятыми бонусами.
type Withdrawals struct {
	Order string    `json:"order"`
	Sum   float32   `json:"sum"`
	Time  time.Time `json:"processed_at"`
}
