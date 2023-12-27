package entity

// Order - Структура, предназначенная для вставки данных в таблицу заказов.
type Order struct {
	Login            string   `bun:"login" json:"-"`
	Order            string   `bun:"order" json:"number"`
	Status           string   `bun:"status" json:"status"`
	UploadedAt       string   `bun:"uploaded_at" json:"uploaded_at"`
	BonusesWithdrawn *float32 `bun:"bonuses_withdrawn"`
	Accrual          *float32 `bun:"accrual" json:"accrual"`
}

// Структура, предназначенная для возврата клиенту данных о заказах с снятыми бонусами
type OWSB struct {
	Order            string   `json:"order"`
	Time             string   `json:"processed_at"`
	BonusesWithdrawn *float32 `json:"sum"`
}

// User отражает информацию о зарегистрированных пользователях
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Points struct designed to receive data from accrual system
type Points struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float32 `json:"accrual"`
}

type Request struct {
	Order
	User
	Points
	OWSB
}
