package entity

type Entity interface {
	GetEntity() *AllEntity
}

// User отражает информацию о зарегистрированных пользователях
type User struct {
	Login     string  `json:"login"`
	Password  string  `json:"password"`
	Balance   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

// Order - Структура, предназначенная для вставки данных в таблицу заказов.
type Orders struct {
	Login            string  `json:"-"`
	Number           string  `json:"number"`
	Status           string  `json:"status"`
	Accrual          float32 `json:"accrual"`
	UploadedAt       string  `json:"uploaded_at"`
	BonusesWithdrawn float32 `json:"order_bonuses_withdrawn"`
}

// cтруктура, предназначенная для возврата клиенту данных о заказах с снятыми бонусами.
type OrdersWithSpentBonuses struct {
	Order            string  `json:"order"`
	Time             string  `json:"processed_at"`
	BonusesWithdrawn float32 `json:"sum"`
}

// cтруктура, предназначенная для получения данных из системы начисления
type OrderUpdateFromAccural struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}

type AllEntity struct {
	User
	Orders
}
