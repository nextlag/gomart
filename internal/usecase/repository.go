package usecase

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/nextlag/gomart/internal/entity"
	"github.com/nextlag/gomart/internal/repository/psql"
	"github.com/nextlag/gomart/pkg/luna"
)

type Storage struct {
	*AllErr
	*psql.Postgres
	*slog.Logger
}

func NewStorage(er *AllErr, db *psql.Postgres, log *slog.Logger) *Storage {
	return &Storage{er, db, log}
}

func (s *Storage) Register(ctx context.Context, login string, password string) error {
	user := &entity.User{
		Login:    login,
		Password: password,
	}
	db := bun.NewDB(s.Postgres.DB, pgdialect.New())

	_, err := db.NewInsert().
		Model(user).
		Exec(ctx)

	if err != nil {
		s.Error("error writing data: ", "error usecase Register", err.Error())
		return err
	}

	return nil
}

func (s *Storage) Auth(ctx context.Context, login, password string) error {
	var user entity.User

	db := bun.NewDB(s.Postgres.DB, pgdialect.New())

	err := db.NewSelect().
		Model(&user).
		Where("login = ? and password = ?", login, password).
		Scan(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) InsertOrder(ctx context.Context, login string, order string) error {
	// Получение текущей даты и времени
	now := time.Now()

	// Инициализация переменной для хранения начисленных бонусов
	bonusesWithdrawn := float32(0)

	// Создание объекта заказа
	userOrder := &entity.Orders{
		Login:            login,
		Number:           order,
		UploadedAt:       now.Format(time.RFC3339),
		Status:           "NEW",
		BonusesWithdrawn: &bonusesWithdrawn,
	}

	// Проверка валидности формата заказа с использованием функции CheckValidOrder
	validOrder := luna.CheckValidOrder(order)
	if !validOrder {
		s.Logger.Debug("InsertOrder", "no valid", validOrder, "status", "invalid order format")
		return s.OrderFormat
	}

	// Создание экземпляра объекта для взаимодействия с базой данных
	db := bun.NewDB(s.DB, pgdialect.New())

	// Поиск заказа в базе данных по номеру заказа
	var checkOrder entity.Orders
	err := db.NewSelect().
		Model(&checkOrder).
		Where(`"order" = ?`, order).
		Scan(ctx)
	if err == nil {
		// Заказ существует
		if checkOrder.Login == login {
			// Заказ принадлежит текущему пользователю
			return s.ThisUser
		}
		// Заказ принадлежит другому пользователю
		return s.AnotherUser
	}

	// Заказ не существует, вставьте его в базу данных
	_, err = db.NewInsert().
		Model(userOrder).
		Exec(ctx)
	if err != nil {
		s.Error("ошибка при записи данных: ", "ошибка в usecase InsertOrder", err.Error())
		return err
	}

	// Возвращение отсутствия ошибок
	return nil
}

func (s *Storage) GetOrders(ctx context.Context, login string) ([]byte, error) {
	// Инициализация среза для хранения заказов
	var allOrders []entity.Orders

	// Создание экземпляра объекта для взаимодействия с базой данных
	db := bun.NewDB(s.Postgres.DB, pgdialect.New())

	// Выполнение SELECT запроса к базе данных для получения заказов по указанному логину
	rows, err := db.NewSelect().
		TableExpr("orders").
		Column("login", "number", "status", "accrual", "uploaded_at", "bonuses_withdrawn").
		Where("login = ?", login).
		Order("uploaded_at ASC").
		Rows(ctx)
	if err != nil {
		s.Logger.Error("ошибка при получении данных", "GetOrders", err.Error())
		return nil, err
	}

	// Проверка наличия ошибок после выполнения запроса
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Итерация по результатам запроса и сканирование данных в структуру Orders
	for rows.Next() {
		var en entity.Orders
		err = rows.Scan(
			&en.Login,
			&en.Number,
			&en.Status,
			&en.Accrual,
			&en.UploadedAt,
			&en.BonusesWithdrawn,
		)
		if err != nil {
			s.Logger.Error("ошибка при сканировании данных", "GetOrders", err.Error())
			return nil, err
		}

		// Добавление полученного заказа в срез allOrders
		allOrders = append(allOrders, en)
	}

	// Преобразование среза заказов в формат JSON
	result, err := json.Marshal(allOrders)
	if err != nil {
		s.Logger.Error("ошибка при маршалинге allOrders", "GetOrders method", err.Error())
		return nil, err
	}

	// Возвращение результата (JSON) и отсутствия ошибок
	return result, nil
}
