package usecase

import (
	"log/slog"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/nextlag/gomart/internal/config"
	"github.com/nextlag/gomart/internal/entity"
)

// GetAccrual - функция, отправляющая ордера в систему начисления
// и получает обновления статуса с начисленными бонусами.
func GetAccrual(order entity.Order, cfg config.Args, log *slog.Logger) entity.Points {
	client := resty.New().SetBaseURL(cfg.Accrual)
	var orderUpdate entity.Points
	for {
		resp, err := client.R().
			SetResult(&orderUpdate).
			Get("/api/orders/" + order.Order)
		if err != nil {
			log.Error("Got error trying to send a get request to accrual: ", err)
			break
		}
		switch resp.StatusCode() {
		case 429:
			time.Sleep(3 * time.Second)
		case 204:
			time.Sleep(1 * time.Second)
		}

		if resp.StatusCode() == 500 {
			log.Error("Internal server error in accrual system: ", err)
			break
		}

		if orderUpdate.Status == "INVALID" || orderUpdate.Status == "PROCESSED" {
			break
		}
	}
	return orderUpdate
}
