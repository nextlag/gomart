package usecase

import (
	"context"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/nextlag/gomart/internal/config"
	"github.com/nextlag/gomart/internal/entity"
)

func GetAccrual(order entity.Orders, cfg config.HTTPServer, log Logger) entity.OrderUpdateFromAccural {
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

func (uc *UseCase) Sync() {
	ticker := time.NewTicker(tick)
	ctx := context.Background()

	for range ticker.C {
		var allOrders []entity.Orders

		order := entity.Orders{}

		db := bun.NewDB(uc.DB, pgdialect.New())

		rows, err := db.NewSelect().
			Model(order).
			Where("status != ? AND status != ?", "PROCESSED", "INVALID").
			Rows(ctx)
		if err != nil {
			return
		}
		err = rows.Err()
		if err != nil {
			return
		}

		for rows.Next() {
			var orderRow entity.Orders
			err = rows.Scan(&orderRow.Users, &orderRow.Number, &orderRow.Status, &orderRow.UploadedAt, &orderRow.BonusesWithdrawn, &orderRow.Accrual)
			if err != nil {
				uc.log.Error("error scanning data", err.Error())
			}
			allOrders = append(allOrders, entity.Orders{
				Number:     orderRow.Number,
				UploadedAt: orderRow.UploadedAt,
				Status:     orderRow.Status,
				Accrual:    orderRow.Accrual,
				Users:      orderRow.Users,
			})
		}
		err = rows.Close()
		if err != nil {
			return
		}

		for _, unfinishedOrder := range allOrders {
			finishedOrder := GetAccrual(unfinishedOrder, uc.cfg, uc.log)
			err = uc.UpdateStatus(ctx, finishedOrder, unfinishedOrder.Users)
			if err != nil {
				return
			}
		}
	}
}

func (uc *UseCase) UpdateStatus(ctx context.Context, orderAccrual entity.OrderUpdateFromAccural, login string) error {

	orderModel := &entity.Orders{}
	userModel := &entity.User{}

	db := bun.NewDB(uc.DB, pgdialect.New())

	_, err := db.NewUpdate().
		Model(orderModel).
		Set("status = ?, accrual = ?", orderAccrual.Status, orderAccrual.Accrual).
		Where(`"number" = ?`, orderAccrual.Order).
		Exec(ctx)
	if err != nil {
		uc.log.Error("error making an update request in order table", err.Error())
		return err
	}

	_, err = db.NewUpdate().
		Model(userModel).
		Set("balance = balance + ?", orderAccrual.Accrual).
		Where(`login = ?`, login).
		Exec(ctx)
	if err != nil {
		uc.log.Error("error making an update request in user table", err.Error())
		return err
	}
	return nil
}
