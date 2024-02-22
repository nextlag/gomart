package controllers

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
)

// PostOrders обрабатывает запрос на создание нового заказа.
//
// Этот метод принимает запрос HTTP POST для создания нового заказа пользователя.
// При успешном выполнении метод возвращает статус Accepted (202) и номер созданного заказа.
// Если в запросе отсутствует тело или происходит ошибка при чтении тела запроса, метод возвращает
// ошибку BadRequest (400) с соответствующим сообщением об ошибке.
// Если происходит ошибка при вставке заказа в базу данных, метод возвращает соответствующий статус и сообщение об ошибке:
//   - если формат заказа неверен, метод возвращает ошибку UnprocessableEntity (422) с сообщением "order format error";
//   - если заказ принадлежит другому пользователю, метод возвращает ошибку Conflict (409) с сообщением "order belongs to another user";
//   - если заказ успешно создан и принадлежит текущему пользователю, метод возвращает статус OK (200) с сообщением "order created";
//   - в остальных случаях метод возвращает ошибку InternalServerError (500) с сообщением "internal server error".
//
// Параметры:
//   - w: http.ResponseWriter - объект для записи HTTP-ответа.
//   - r: *http.Request - объект HTTP-запроса.
//
// Возвращаемые значения:
//   - нет.
func (c *Controller) PostOrders(w http.ResponseWriter, r *http.Request) {
	// Получаем объект ошибки из UseCase
	er := c.uc.Do().Err()
	// Получаем логин пользователя из контекста запроса
	user, _ := r.Context().Value(auth.LoginKey).(string)

	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	order := string(body)
	switch {
	case order == "":
		// Если тело запроса отсутствует, возвращаем ошибку BadRequest (400)
		http.Error(w, er.ErrRequestFormat.Error(), http.StatusBadRequest)
		return
	case err != nil:
		// Если произошла ошибка при чтении тела запроса, возвращаем ошибку InternalServerError (500)
		c.log.Error("body reading error", "error PostOrder handler", err.Error())
		http.Error(w, er.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	// Вставляем заказ в базу данных
	err = c.uc.DoInsertOrder(r.Context(), user, order)
	switch {
	case errors.Is(err, er.ErrOrderFormat):
		// Если формат заказа неверен, возвращаем ошибку UnprocessableEntity (422)
		c.log.Error("insert Order 422", "error", err.Error())
		http.Error(w, er.ErrOrderFormat.Error(), http.StatusUnprocessableEntity)
		return
	case errors.Is(err, er.ErrAnotherUser):
		// Если заказ принадлежит другому пользователю, возвращаем ошибку Conflict (409)
		c.log.Error("insert Order 409", "error", err.Error())
		http.Error(w, er.ErrAnotherUser.Error(), http.StatusConflict)
		return
	case errors.Is(err, er.ErrThisUser):
		// Если заказ успешно создан и принадлежит текущему пользователю, возвращаем статус OK (200)
		c.log.Error("insert Order 200", "error", err.Error())
		http.Error(w, er.ErrThisUser.Error(), http.StatusOK)
		return
	default:
		// В остальных случаях возвращаем ошибку InternalServerError (500)
		http.Error(w, er.ErrOrderAccepted.Error(), http.StatusAccepted)
	}

	// Логируем успешное получение заказа
	c.log.Info("order received", "user", user, "order", order)
	o := fmt.Sprintf("order number: %s\n", order)
	w.Write([]byte(o))
}
