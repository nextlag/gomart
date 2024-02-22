// Package usecase provides business logic for interacting with the database and external service.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/nextlag/gomart/internal/config"
	"github.com/nextlag/gomart/internal/entity"
)

// OrderResponse - структура предназначена для получения данных из системы начисления бонусов.
type OrderResponse struct {
	Order   string  `json:"order"`   // Номер заказа
	Status  string  `json:"status"`  // Статус заказа
	Accrual float32 `json:"accrual"` // Сумма начисления бонусов
}

// GetAccrual - функция отправляет HTTP-запрос и возвращает структуру OrderResponse.
// Функция получает данные заказа из системы начисления.
// Функция принимает данные заказа order типа entity.Order и канал stop для получения сигнала об остановке выполнения запроса.
// Возвращает структуру OrderResponse с данными о статусе заказа и ошибку.
// Функция выполняет цикл запросов к системе начисления бонусов до получения сигнала об остановке или изменения статуса заказа на "INVALID" или "PROCESSED".
// В случае получения сигнала об остановке, функция завершает выполнение и возвращает пустую структуру OrderResponse и nil.
// Если при выполнении запроса произошла ошибка, функция возвращает эту ошибку.
// При получении статуса "PROCESSING" функция ожидает 1 секунду и повторяет запрос.
// При получении статуса "429" (слишком много запросов) функция завершает выполнение и возвращает пустую структуру OrderResponse и ошибку.
// При получении статуса "204" (нет содержимого) функция завершает выполнение и возвращает пустую структуру OrderResponse и ошибку.
// При получении статуса "500" (внутренняя ошибка сервера начисления бонусов) функция завершает выполнение и возвращает пустую структуру OrderResponse и ошибку.
func GetAccrual(order entity.Order, stop chan struct{}) (OrderResponse, error) {
	client := resty.New().SetBaseURL(config.Cfg.Accrual)
	var orderUpdate OrderResponse

	// Выполняем цикл, пока не получим сигнал остановки из канала stop
	for {
		select {
		case <-stop:
			return orderUpdate, nil
		default:
			resp, err := client.R().
				SetResult(&orderUpdate).
				Get("/api/orders/" + order.Order)

			if err != nil {
				log.Printf("got error trying to send a get request to accrual: %v", err)
				return orderUpdate, err
			}
			log.Printf("response body: %s", resp.String())

			switch resp.StatusCode() {
			case 200:
				switch orderUpdate.Status {
				case "INVALID", "PROCESSED":
					log.Printf("Exiting the loop. Order status: %s", orderUpdate.Status)
					return orderUpdate, nil
				case "PROCESSING":
					log.Printf("Order status is PROCESSING. Sleeping for 1 second before the next request.")
					time.Sleep(1 * time.Second)
				default:
					log.Printf("Unknown order status: %s. Sleeping for 1 second before the next request.", orderUpdate.Status)
					time.Sleep(1 * time.Second)
				}
			case 429:
				return orderUpdate, errors.New("request limit exceeded")
			case 204:
				return orderUpdate, errors.New("order isn't registered")
			case 500:
				log.Printf("internal server error in accrual system: %v", err)
				return orderUpdate, fmt.Errorf("internal server error in accrual system: %v", err)
			default:
				log.Printf("Unexpected status code: %d. Sleeping for 1 second before the next request.", resp.StatusCode())
				time.Sleep(1 * time.Second)
			}
		}
	}
}

// Sync function for synchronizing orders.
// The function periodically checks the status of orders and updates them in the database.
func (uc *UseCase) Sync(stop chan struct{}) error {
	ticker := time.NewTicker(tick)
	ctx := context.Background()

	for range ticker.C {
		var allOrders []entity.Order
		order := &entity.Order{}
		db := bun.NewDB(uc.DB, pgdialect.New())

		rows, err := db.NewSelect().
			Model(order).
			Where("status != ? AND status != ?", "PROCESSED", "INVALID").
			Rows(ctx)
		rows.Err()
		if err != nil {
			return err
		}

		for rows.Next() {
			var orderRow entity.Order
			err = rows.Scan(&orderRow.Users, &orderRow.Order, &orderRow.Status, &orderRow.Accrual, &orderRow.UploadedAt, &orderRow.BonusesWithdrawn)
			if err != nil {
				return err
			}
			allOrders = append(allOrders, entity.Order{
				Users:      orderRow.Users,
				Order:      orderRow.Order,
				Status:     orderRow.Status,
				Accrual:    orderRow.Accrual,
				UploadedAt: orderRow.UploadedAt,
			})
		}
		err = rows.Close()
		if err != nil {
			return err
		}

		for _, unfinishedOrder := range allOrders {
			finishedOrder, err := GetAccrual(unfinishedOrder, stop)
			if err != nil {
				return err
			}
			log.Print("finished", finishedOrder)
			err = uc.UpdateStatus(ctx, finishedOrder, unfinishedOrder.Users)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// UpdateStatus function to update the order status and user balance in the database.
// The function accepts an OrderResponse structure with updated order data and user login.
func (uc *UseCase) UpdateStatus(ctx context.Context, orderAccrual OrderResponse, login string) error {

	orderModel := &entity.Order{}
	userModel := &entity.User{}

	db := bun.NewDB(uc.DB, pgdialect.New())

	_, err := db.NewUpdate().
		Model(orderModel).
		Set("status = ?, accrual = ?", orderAccrual.Status, orderAccrual.Accrual).
		Where(`"order" = ?`, orderAccrual.Order).
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = db.NewUpdate().
		Model(userModel).
		Set("balance = balance + ?", orderAccrual.Accrual).
		Where(`login = ?`, login).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("error making an update request in user table: %v", err)
	}
	return nil
}
