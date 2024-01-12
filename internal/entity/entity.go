package entity

type Entity interface {
	GetEntity() *AllEntity
}

// Order - Структура, предназначенная для вставки данных в таблицу заказов.
type Orders struct {
	Login            string  `bun:"type:varchar(255)"`
	Order            string  `bun:"type:varchar(255),unique"`
	Status           string  `bun:"type:varchar(255)"`
	UploadedAt       string  `bun:"type:timestamp"`
	BonusesWithdrawn float32 `bun:"type:float"`
	Accrual          float32 `bun:"type:float"`
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
