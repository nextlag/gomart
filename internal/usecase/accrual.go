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

// batchSize - размер батча - при большом количестве скопившихся заказов (или при горизонтальном масштабировании)
// функция не попытается захватить сразу все записи из базы,
// не затупит и не пропустит следующий tick.
const batchSize = 30

// OrderResponse - структура предназначена для получения данных из системы начисления бонусов.
type OrderResponse struct {
	Order   string  `json:"order"`   // Номер заказа
	Status  string  `json:"status"`  // Статус заказа
	Accrual float32 `json:"accrual"` // Сумма начисления бонусов
}

// GetAccrual выполняет запрос к системе начисления для получения информации о заказе.
// Функция отправляет HTTP GET запрос к указанному API для получения статуса заказа.
// Параметры:
//   - order: структура, содержащая информацию о заказе, включая его номер.
//   - stop: канал для получения сигнала остановки выполнения функции.
//
// Возвращаемые значения:
//   - OrderResponse: структура, содержащая информацию о статусе заказа, начислении и других данных.
//   - error: ошибка, возникающая в случае невозможности выполнения запроса или получения данных.
//
// Функция выполняет цикл запросов до получения сигнала остановки из канала stop или успешного завершения запроса.
// При получении статуса 200 функция проверяет статус заказа из ответа. Если заказ
// имеет статус "INVALID" или "PROCESSED", цикл завершается и возвращается информация о заказе.
// Если статус заказа "PROCESSING", функция приостанавливает выполнение на 1 секунду перед
// следующим запросом. При получении статуса 429 (слишком много запросов), функция возвращает
// ошибку "request limit exceeded". При статусе 204 (заказ не зарегистрирован), возвращается
// ошибка "order isn't registered". При получении статуса 500 (внутренняя ошибка сервера
// в системе начислений), возвращается ошибка "internal server error in accrual system".
// Если выполнение функции завершается по сигналу остановки, она возвращает текущее состояние
// заказа без ошибки.
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
				return orderUpdate, errors.New("internal server error in accrual system")
			}
		}
	}
}

// Sync выполняет синхронизацию заказов с системой начисления бонусов.
// Функция периодически запрашивает статусы незавершенных заказов и обновляет их статусы
// в базе данных в соответствии с полученной информацией.
// Параметры:
//   - stop: канал для получения сигнала остановки выполнения функции.
//
// Возвращаемое значение:
//   - error: в случае возникновения ошибки при выполнении синхронизации.
//
// Функция создает новый тикер, который запускает процесс синхронизации через определенные интервалы времени.
// При каждом срабатывании тикера, функция выполняет запрос к базе данных для получения списка незавершенных заказов.
// Затем она запускает цикл обработки этих заказов, вызывая функцию GetAccrual для каждого заказа
// и обновляя статусы заказов в базе данных согласно полученной информации. Функция продолжает
// работу до получения сигнала остановки из канала stop.
func (uc *UseCase) Sync(stop chan struct{}) error {
	ticker := time.NewTicker(tick)
	defer ticker.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		select {
		case <-ticker.C:
			var allOrders []entity.Order
			order := &entity.Order{}
			db := bun.NewDB(uc.DB, pgdialect.New())

			rows, err := db.NewSelect().
				Model(order).
				Where("status != ? AND status != ?", "PROCESSED", "INVALID").
				Limit(batchSize).
				Rows(ctx)
			if err != nil {
				log.Printf("error selecting orders: %v", err)
				continue
			}

			for rows.Next() {
				var orderRow entity.Order
				err = rows.Scan(&orderRow.UserName, &orderRow.Order, &orderRow.Status, &orderRow.Accrual, &orderRow.UploadedAt, &orderRow.BonusesWithdrawn)
				if err != nil {
					log.Printf("error scanning row: %v", err)
					continue
				}

				allOrders = append(allOrders, entity.Order{
					UserName:   orderRow.UserName,
					Order:      orderRow.Order,
					Status:     orderRow.Status,
					Accrual:    orderRow.Accrual,
					UploadedAt: orderRow.UploadedAt,
				})
			}
			err = rows.Close()
			if err != nil {
				log.Printf("error closing rows: %v", err)
				continue
			}

			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				log.Printf("error beginning transaction: %v", err)
				continue
			}

			for _, unfinishedOrder := range allOrders {
				select {
				case <-stop:
					cancel()
					return nil // В случае получения сигнала остановки, завершаем выполнение без ошибок
				default:
					finishedOrder, err := GetAccrual(unfinishedOrder, stop)
					if err != nil {
						log.Printf("error getting accrual: %v", err)
						tx.Rollback()
						continue // Пропустить текущую итерацию цикла и перейти к следующей итерации
					}
					log.Print("finished", finishedOrder)
					err = uc.UpdateStatus(ctx, finishedOrder, unfinishedOrder.UserName, tx)
					if err != nil {
						log.Printf("error updating status: %v", err)
						tx.Rollback()
						continue // Пропустить текущую итерацию цикла и перейти к следующей итерации
					}
				}
			}

			if err = tx.Commit(); err != nil {
				log.Printf("error committing transaction: %v", err)
				continue
			}
		case <-stop:
			cancel()
			return nil // В случае получения сигнала остановки, завершаем выполнение без ошибок
		}
	}
}

// UpdateStatus обновляет статус заказа и баланс пользователя в базе данных на основе полученных данных о начислении.
// Функция принимает контекст ctx типа context.Context, структуру OrderResponse с информацией о начислении,
// и логин пользователя login.
// Параметры:
//   - ctx: контекст выполнения запроса.
//   - orderAccrual: структура OrderResponse с информацией о статусе заказа и начислении.
//   - login: логин пользователя, чей баланс нужно обновить.
//
// Возвращаемое значение:
//   - error: в случае возникновения ошибки при выполнении запроса к базе данных.
//
// Функция выполняет два отдельных запроса к базе данных для обновления статуса заказа и баланса пользователя.
// Сначала она обновляет статус заказа и начисление в таблице заказов, а затем обновляет баланс пользователя
// в соответствии с начисленной суммой. Если произошла ошибка при выполнении запросов к базе данных,
// функция возвращает ошибку.
func (uc *UseCase) UpdateStatus(ctx context.Context, orderAccrual OrderResponse, login string, tx bun.Tx) error {

	orderModel := &entity.Order{}
	userModel := &entity.User{}

	// Используем tx для создания запроса обновления
	_, err := tx.NewUpdate().
		Model(orderModel).
		Set("status = ?, accrual = ?", orderAccrual.Status, orderAccrual.Accrual).
		Where(`"order" = ?`, orderAccrual.Order).
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = tx.NewUpdate().
		Model(userModel).
		Set("balance = balance + ?", orderAccrual.Accrual).
		Where(`login = ?`, login).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("error making an update request in user table: %v", err)
	}
	return nil
}
