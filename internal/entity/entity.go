package entity

// User - отражает информацию о зарегистрированных пользователях
type User struct {
	Login     string  `json:"login"`
	Password  string  `json:"password"`
	Balance   float32 `json:"balance"`
	Withdrawn float32 `json:"withdrawn"`
}

// Orders - cтруктура, предназначенная для вставки данных в таблицу заказов.
type Orders struct {
	Users            string  `json:"users,omitempty"`
	Number           string  `json:"number"`
	Status           string  `json:"status"`
	Accrual          float32 `json:"accrual,omitempty"`
	UploadedAt       string  `json:"uploaded_at"`
	BonusesWithdrawn float32 `json:"bonuses_withdrawn,omitempty"`
}

type AllEntity struct {
	*User
	*Orders
}

// Withdrawals - cтруктура, предназначенная для возврата клиенту данных о заказах с снятыми бонусами.
type Withdrawals struct {
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
