# Accrual System API Server

### Introduction

Этот сервер API взаимодействует с системой начисления, обеспечивая функциональные возможности, связанные с балансом
пользователя, аутентификацией и управлением заказами. Он использует PostgreSQL для хранения данных и платформу Chi для
маршрутизации и обслуживания запросов.
Сервер включает в себя конечные точки для получения баланса пользователя, вывода бонусов, регистрации и аутентификации
пользователей, а также загрузки и получения заказов.
Кроме того, на сервере имеется промежуточное программное обеспечение для проверки аутентификации пользователя с помощью
файлов cookie и сжатия данных.

### Флаги запуска

1. **-a** _сокет приложения (по умолчанию :8080)_
2. **-d** _реквизиты доступа к бд_
3. **-k** _установка секретного ключа для токена_
4. **-l** _уровень логирования: info, debug, wrong, error, (по умолчанию info)_
5. **-r** _сокет системы расчета начисления бонусов(по умолчанию :8081)_

### Balance

1. **GET** /user/balance - _получение баланса пользователя, включая снятую сумму_
2. **POST** /user/balance/withdraw - _вывод бонусов пользователей_

### Auth

1. **POST** /user/register - _регистрация и аутентификация пользователя_
2. **POST** /user/login - _аутентификация пользователя и установка файла cookie аутентификации_

### Order

1. **POST** /user/orders - _загрузка заказа на сервер_
2. **GET** /user/orders - _получение заказов пользователей_
3. **GET** /user/withdrawals - _получение заказов с потраченными бонусами_

## Project Structure

Описание директорий и файлов проекта

- **cmd**
    - **accrual** - _содержит двоичный файл системы начисления_
    - **gophermart** - _содержит основной пакет_
        - main.go - _основная функция_
- **internal**
    - **config**
        - config.go - _функции и структуры настройки конфигурации_
        - loglevel.go - _определяет пользовательский тип LogLevelValue и реализует интерфейс flag.Value для него_
    - **controllers** - _слой обработчиков запросов_
        - **mosck**
            - mocsk.go - _mocks слоя обработчика запросов_
        - authentication.go - _аутентификация пользователя_
        - balance.go - _получение текущего баланса, счёта, баллов лояльности пользователя_
        - controllers.go - _содержит обработчики запросов для API_
        - controllers_test.go - _тесты хендлеров_
        - get_orders.go - _получение списка загруженных пользователем номеров заказов, статусов их обработки и
          информации о начислениях_
        - post_orders.go - _загрузка пользователем номера заказа для расчёта_
        - register.go - _регистрация пользователя_
        - withdraw.go - _запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа_
        - withdrawals.go - _получение информации о выводе средств с накопительного счёта пользователем_
    - **entity** - _слой структур бизнес-логики_
        - entity.go - _основные структуры бизнес-логики_
    - **mw** - _middleware_
        - **auth**
            - auth.go - _пакет получения токена аутентификации_
            - cookie.go - _middleware аутентификации_
        - **gzip**
            - compress.go - _пакет gzip, который обеспечивает сжатие и распаковку данных в формате gzip для
              HTTP-запросов и ответов_
            - gzip.go - _middleware gzip_
        - **logger**
            - slogger.go - _middleware логгера_
    - **repository**
        - **psql**
            - psql.go _функция инициализации базы данных postgres_
    - **usecase** _слой бизнес-логики_.
        - accrual.go - _взаимодействие с системой расчёта начислений баллов лояльности_
        - errors.go - _ошибки_
        - mocks.go - _mocks пакета usecase_
        - repository.go - _бизнес-логика приложения_
        - storage.go - _функции для работы с базой данных_
        - usecase.go - _основной пакет usecase, содержащий интерфейс и структуру, представляющую бизнес-логику
          приложения_
- **pkg**
    - **generatestring**
        - generatestring.go - _генератор строк_
    - **logger**
        - **slogpretty**
            - slogpretty.go - _обертка логгера_
    - **luna**
        - luna.go - _проверка валидности номера заказа алгоритмом 'Луна'_

