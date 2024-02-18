// Package usecase provides business logic for interacting with the database and external service.
package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/nextlag/gomart/internal/config"
	"github.com/nextlag/gomart/internal/entity"
)

// OrderResponse structure is designed to receive data from the accrual system
type OrderResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}

// GetAccrual is a function that sends an HTTP request and returns an OrderResponse structure.
// The function receives order data from the accrual system.
func GetAccrual(order entity.Orders, stop chan struct{}) (OrderResponse, error) {
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
				log.Println("429 status code. Sleeping for 3 seconds.")
				time.Sleep(3 * time.Second)
			case 204:
				log.Println("204 status code. Sleeping for 1 second.")
				time.Sleep(1 * time.Second)
			case 500:
				log.Printf("internal server error in accrual system: %v", err)
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
		var allOrders []entity.Orders
		order := &entity.Orders{}
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
			var orderRow entity.Orders
			err = rows.Scan(&orderRow.Users, &orderRow.Order, &orderRow.Status, &orderRow.Accrual, &orderRow.UploadedAt, &orderRow.BonusesWithdrawn)
			if err != nil {
				return err
			}
			allOrders = append(allOrders, entity.Orders{
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

	orderModel := &entity.Orders{}
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
		return fmt.Errorf("error making an update request in user table: %v\n", err)
	}
	return nil
}
