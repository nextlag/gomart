package controllers

import (
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
)

// GetOrders обрабатывает запрос на получение заказов пользователя.
//
// Этот метод принимает запрос HTTP GET для получения списка заказов пользователя.
// При успешном выполнении метод возвращает список заказов пользователя в формате JSON и статус OK (200).
// Если происходит ошибка при получении списка заказов из UseCase, метод возвращает ошибку InternalServerError (500)
// с сообщением "internal server error".
//
// Параметры:
//   - w: http.ResponseWriter - объект для записи HTTP-ответа.
//   - r: *http.Request - объект HTTP-запроса.
//
// Возвращаемые значения:
//   - нет.
func (c *Controller) GetOrders(w http.ResponseWriter, r *http.Request) {
	// Получаем логин пользователя из контекста запроса
	user, _ := r.Context().Value(auth.LoginKey).(string)
	// Получаем объект ошибки из UseCase
	er := c.uc.Do().Err()

	// Получаем список заказов пользователя из UseCase
	result, err := c.uc.DoGetOrders(r.Context(), user)
	if err != nil {
		// Если произошла ошибка при получении списка заказов, логируем её и возвращаем ошибку InternalServerError
		c.log.Error("handler GetOrders", "error", err.Error())
		http.Error(w, er.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовок Content-Type и код статуса OK (200) в ответе
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Возвращаем JSON с заказами пользователя
	w.Write(result)
}
