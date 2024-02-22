package controllers

import (
	"errors"
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
	"github.com/nextlag/gomart/pkg/logger/l"
)

// Withdrawals обрабатывает запрос на получение истории списаний средств пользователя.
//
// Этот метод принимает запрос HTTP GET для получения истории списаний средств пользователя.
// При успешном выполнении метод возвращает историю списаний в формате JSON и статус OK (200).
// Если история списаний пуста, метод возвращает статус NoContent (204).
// Если происходит ошибка при получении истории списаний из UseCase, метод возвращает ошибку InternalServerError (500)
// с соответствующим сообщением об ошибке.
//
// Параметры:
//   - w: http.ResponseWriter - объект для записи HTTP-ответа.
//   - r: *http.Request - объект HTTP-запроса.
//
// Возвращаемые значения:
//   - нет.
func (c *Controller) Withdrawals(w http.ResponseWriter, r *http.Request) {
	log := l.L(c.ctx)
	// Получаем объект ошибки из UseCase
	er := c.uc.Do().Err()
	// Получаем логин пользователя из контекста
	user, _ := r.Context().Value(auth.LoginKey).(string)

	// Получаем историю списаний средств пользователя из UseCase
	result, err := c.uc.DoGetWithdrawals(c.ctx, user)
	switch {
	case errors.Is(err, er.ErrNoRows):
		// Если история списаний пуста, возвращаем статус NoContent (204)
		log.Error("withdrawals handler", l.ErrAttr(err))
		http.Error(w, "there is no write off", http.StatusNoContent)
		return
	case err != nil:
		// Если произошла ошибка при получении истории списаний, возвращаем ошибку InternalServerError (500)
		log.Error("withdrawals handler", l.ErrAttr(err))
		http.Error(w, er.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовок Content-Type и код статуса OK (200) в ответе
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Возвращаем JSON с историей списаний средств пользователя
	w.Write(result)
}
