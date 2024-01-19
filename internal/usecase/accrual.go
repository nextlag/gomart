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

func GetAccrual(order entity.Orders, cfg config.HTTPServer, log Logger) entity.OrderUpdateFromAccrual {
	client := resty.New().SetBaseURL(cfg.Accrual)
	var orderUpdate entity.OrderUpdateFromAccrual
	for {
		resp, err := client.R().
			SetResult(&orderUpdate).
			Get("/api/orders/" + order.Number)
		if err != nil {
			break
		}
		switch resp.StatusCode() {
		case 429:
			time.Sleep(3 * time.Second)
		case 204:
			time.Sleep(1 * time.Second)
		}

		if resp.StatusCode() == 500 {
			break
		}

		if orderUpdate.Status == "INVALID" || orderUpdate.Status == "PROCESSED" {
			break
		}
	}
	return orderUpdate
}

func (uc *UseCase) Sync() error {
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
			err = rows.Scan(&orderRow.Users, &orderRow.Number, &orderRow.Status, &orderRow.Accrual, &orderRow.UploadedAt, &orderRow.BonusesWithdrawn)
			if err != nil {
				return err
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
			return err
		}

		for _, unfinishedOrder := range allOrders {
			finishedOrder := GetAccrual(unfinishedOrder, uc.cfg, uc.log)
			err = uc.UpdateStatus(ctx, finishedOrder, unfinishedOrder.Users)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (uc *UseCase) UpdateStatus(ctx context.Context, orderAccrual entity.OrderUpdateFromAccrual, login string) error {

	orderModel := &entity.Orders{}
	userModel := &entity.User{}

	db := bun.NewDB(uc.DB, pgdialect.New())

	_, err := db.NewUpdate().
		Model(orderModel).
		Set("status = ?, accrual = ?", orderAccrual.Status, orderAccrual.Accrual).
		Where(`"number" = ?`, orderAccrual.Number).
		Exec(ctx)
	if err != nil {
		uc.log.Error("error making an update request in order table", err)
		return err
	}

	_, err = db.NewUpdate().
		Model(userModel).
		Set("balance = balance + ?", orderAccrual.Accrual).
		Where(`login = ?`, login).
		Exec(ctx)
	if err != nil {
		uc.log.Error("error making an update request in user table", err)
		return err
	}
	return nil
}
