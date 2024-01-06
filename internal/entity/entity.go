package entity

// Order - Структура, предназначенная для вставки данных в таблицу заказов.
type Order struct {
	Login            string   `bun:"login" json:"-"`
	Order            string   `bun:"order" json:"order"`
	Status           string   `bun:"status" json:"order_status"`
	UploadedAt       string   `bun:"uploaded_at" json:"uploaded_at"`
	BonusesWithdrawn *float32 `bun:"bonuses_withdrawn"`
	Accrual          *float32 `bun:"accrual" json:"order_accrual"`
}

// Структура, предназначенная для возврата клиенту данных о заказах с снятыми бонусами
type OWSB struct {
	Order            string   `json:"owsb_order"`
	Time             string   `json:"processed_at"`
	BonusesWithdrawn *float32 `json:"sum"`
}

// User отражает информацию о зарегистрированных пользователях
type User struct {
	Login     string  `json:"login"`
	Password  string  `json:"password"`
	Balance   float32 `json:"balance"`
	Withdrawn float32 `json:"withdrawn"`
}

// Points struct designed to receive data from accrual system
type Points struct {
	Order   string   `json:"points_order"`
	Status  string   `json:"points_status"`
	Accrual *float32 `json:"points_accrual"`
}

type Entity struct {
	Order
	User
	Points
	OWSB
}
