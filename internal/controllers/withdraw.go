package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/nextlag/gomart/internal/mw/auth"
	"github.com/nextlag/gomart/pkg/logger/l"
)

// debit - структура используемая для анализа json-запроса на вывод бонусов при оформлении заказа.
type debit struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}

// Withdraw обрабатывает запрос на списание средств со счета пользователя.
//
// Этот метод принимает запрос HTTP POST с JSON-данными, содержащими номер заказа и сумму для списания.
// При успешном списании средств метод возвращает статус OK (200).
// Если указанный пользователь не имеет достаточного баланса для списания, метод возвращает ошибку PaymentRequired (402).
// Если происходит ошибка при декодировании JSON или при списании средств, метод возвращает ошибку InternalServerError (500)
// с соответствующим сообщением об ошибке.
//
// Параметры:
//   - w: http.ResponseWriter - объект для записи HTTP-ответа.
//   - r: *http.Request - объект HTTP-запроса.
//
// Возвращаемые значения:
//   - нет.
func (c *Controller) Withdraw(w http.ResponseWriter, r *http.Request) {
	log := l.L(c.ctx)
	// Получаем объект ошибки из UseCase
	er := c.uc.Do().Err()
	// Получаем логин пользователя из контекста
	user, _ := r.Context().Value(auth.LoginKey).(string)

	// Декодируем JSON-данные из тела запроса в структуру debit
	var request debit
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}
	// Логируем запрос на списание средств
	log.Debug("debet request", "user", user, "order", request.Order, "sum", request.Sum)

	// Вызываем метод DoDebit UseCase для списания средств
	err := c.uc.Do().DoDebit(r.Context(), user, request.Order, request.Sum)
	switch {
	case errors.Is(err, er.ErrNoBalance):
		// Если недостаточно средств на счете, возвращаем ошибку PaymentRequired (402)
		log.Error("there are insufficient funds in the account", l.ErrAttr(err))
		http.Error(w, er.ErrNoBalance.Error(), http.StatusPaymentRequired)
		return
	case errors.Is(err, er.ErrOrderFormat):
		// Если неверный формат заказа, возвращаем ошибку UnprocessableEntity (422)
		log.Error("withdraw OrderFormat", l.ErrAttr(err))
		http.Error(w, er.ErrOrderFormat.Error(), http.StatusUnprocessableEntity)
		return
	case errors.Is(err, er.ErrThisUser) || errors.Is(err, er.ErrAnotherUser):
		// Если заказ уже обработан, возвращаем ошибку Conflict (409)
		log.Debug("withdraw", "user", user, "order", request.Order)
		log.Error("withdraw AnotherUser", l.ErrAttr(err))
		http.Error(w, "order is already loaded", http.StatusConflict)
		return
	case err != nil:
		// Если произошла другая ошибка при списании средств, возвращаем ошибку InternalServerError (500)
		log.Error("withdraw handler", l.ErrAttr(err))
		http.Error(w, er.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}
	// Возвращаем успешный статус и сообщение об успешном списании средств
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("bonuses were written off success"))
}
