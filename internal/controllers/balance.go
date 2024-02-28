package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
	"github.com/nextlag/gomart/pkg/logger/l"
)

type userBalance struct {
	Balance   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

// Balance обрабатывает запрос на получение баланса пользователя.
//
// Этот метод принимает запрос HTTP GET для получения баланса пользователя.
// При успешном выполнении метод возвращает текущий баланс и сумму снятых средств пользователя в формате JSON
// и статус OK (200).
// Если происходит ошибка при получении баланса из UseCase, метод возвращает ошибку InternalServerError (500)
// с сообщением "error get balance".
//
// Параметры:
//   - w: http.ResponseWriter - объект для записи HTTP-ответа.
//   - r: *http.Request - объект HTTP-запроса.
//
// Возвращаемые значения:
//   - нет.
func (c *Controller) Balance(w http.ResponseWriter, r *http.Request) {
	log := l.L(c.ctx)
	// Получаем логин пользователя из контекста запроса
	login, _ := r.Context().Value(auth.LoginKey).(string)
	// Получаем текущий баланс и сумму снятых средств пользователя из UseCase
	balance, withdrawn, err := c.uc.DoGetBalance(c.ctx, login)
	if err != nil {
		// Если произошла ошибка при получении баланса, логируем её и возвращаем ошибку InternalServerError
		log.Error("Balance handler", "balance", balance, "withdrawn", withdrawn, l.ErrAttr(err))
		http.Error(w, "error get balance", http.StatusInternalServerError)
		return
	}

	// Формируем структуру с текущим балансом и суммой снятых средств
	user := userBalance{
		Balance:   balance,
		Withdrawn: withdrawn,
	}

	// Преобразуем структуру в JSON
	result, err := json.Marshal(user)
	if err != nil {
		// Если произошла ошибка при маршалинге в JSON, возвращаем пустой ответ
		return
	}

	// Логируем успешное получение баланса
	log.Info("GetBalance handler", "balance", balance, "withdrawn", withdrawn)

	// Устанавливаем заголовок Content-Type и код статуса OK (200) в ответе
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Возвращаем JSON с балансом пользователя
	w.Write(result)
}
