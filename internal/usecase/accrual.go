package usecase

import (
	"log/slog"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/nextlag/gomart/internal/config"
	"github.com/nextlag/gomart/internal/entity"
)

func GetAccrual(order entity.Orders, cfg config.HTTPServer, log *slog.Logger) entity.OrderUpdateFromAccural {
	client := resty.New().SetBaseURL(cfg.Accrual)
	var orderUpdate entity.OrderUpdateFromAccural

	maxWaitTime := 5 * time.Minute  // Максимальное время ожидания
	waitInterval := 3 * time.Second // Интервал ожидания

	startTime := time.Now()

	for elapsed := time.Since(startTime); elapsed < maxWaitTime; elapsed = time.Since(startTime) {
		resp, err := client.R().
			SetResult(&orderUpdate).
			Get("/api/orders/" + order.Number)

		if err != nil {
			log.Error("error when sending a GET request to the accrual system:", "error GetAccrual", err.Error())
			time.Sleep(waitInterval)
			continue
		}

		switch resp.StatusCode() {
		case 429:
			// Пауза 3 секунды при получении кода 429 (слишком много запросов)
			time.Sleep(waitInterval)
			continue
		case 204:
			// Пауза 1 секунда при получении кода 204 (успешный запрос, но нет данных)
			time.Sleep(waitInterval)
			continue
		}

		if resp.StatusCode() == 500 {
			log.Error("internal server error in the accrual system:", nil)
			time.Sleep(waitInterval)
			continue
		}

		if orderUpdate.Status == "INVALID" || orderUpdate.Status == "PROCESSED" {
			return orderUpdate
		}

		// Пауза перед следующей попыткой
		time.Sleep(waitInterval)
	}

	return orderUpdate
}

// // GetAccrual отправляет ордера в систему начисления и получает обновления статуса с начисленными бонусами.
// func GetAccrual(order entity.Order, cfg config.HTTPServer, log *slog.Logger) entity.Points {
// 	// Инициализация HTTP-клиента с базовым URL из конфигурации
// 	client := resty.New().SetBaseURL(cfg.Accrual)
//
// 	// Переменная для хранения обновлений статуса ордера
// 	var orderUpdate entity.Points
//
// 	// Бесконечный цикл для повторных запросов до получения ожидаемого результата
// 	for {
// 		// Отправка GET-запроса к системе начисления с номером ордера
// 		resp, err := client.R().
// 			SetResult(&orderUpdate).
// 			Get("/api/orders/" + order.Order)
//
// 		// Проверка ошибок при отправке запроса
// 		if err != nil {
// 			log.Error("Ошибка при отправке GET-запроса в систему начисления: ", err)
// 			break
// 		}
//
// 		// Обработка различных статусов ответа
// 		switch resp.StatusCode() {
// 		case 429:
// 			// Пауза 3 секунды при получении кода 429 (слишком много запросов)
// 			time.Sleep(3 * time.Second)
// 		case 204:
// 			// Пауза 1 секунда при получении кода 204 (успешный запрос, но нет данных)
// 			time.Sleep(1 * time.Second)
// 		}
//
// 		// Проверка на внутреннюю ошибку сервера
// 		if resp.StatusCode() == 500 {
// 			log.Error("Внутренняя ошибка сервера в системе начисления: ", err)
// 			break
// 		}
//
// 		// Проверка статуса обновления ордера, если статус "INVALID" или "PROCESSED", выход из цикла
// 		if orderUpdate.Status == "INVALID" || orderUpdate.Status == "PROCESSED" {
// 			break
// 		}
// 	}
//
// 	// Возвращение обновленного статуса ордера
// 	return orderUpdate
// }
