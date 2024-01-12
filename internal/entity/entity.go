package entity

type Entity interface {
	GetEntity() *AllEntity
}

// Order - Структура, предназначенная для вставки данных в таблицу заказов.
type Orders struct {
	Login            string   `json:"-"`
	Number           string   `json:"number"`
	Status           string   `json:"status"`
	Accrual          *float32 `json:"accrual"`
	UploadedAt       string   `json:"uploaded_at"`
	BonusesWithdrawn *float32 `json:"order_bonuses_withdrawn"`
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

type AllEntity struct {
	Orders
	User
	Points
	OWSB
}
