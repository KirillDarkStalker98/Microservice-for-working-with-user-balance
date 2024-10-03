# Microservice-for-working-with-user-balance (Микросервис для работы с балансом пользователей)

## Функционал API

* Создание пользователя и добавление ему имени

* Зачисление средств

* Получение баланса пользователя 

* Покупка услуг и списание средств (зачисление средств на резервный счёт перед покупкой и списание средств с него)

* Перевод средств от пользователя к пользователю 

* Создание, изменение и удаление услуг которые можно купить

* Получение месячного отчёта 

* Просмотр транзакций всех пользователей или одного пользователя. 

Представляет из себя HTTP API с форматом JSON как при отправке запроса, так и при получении результата.


## Что нужно для запуска на windows:

1. PostgreSQL
							
2. Язык Golang
	
3. Система очередей Redis (запуск через WSL)

* В начале для работы с сервисом в PostgreSQL необходимо создать базу данных с названием "service_db", затем выбрать базу, которая была создана, и прямо в консоль вставить SQL код из файла "bd.sql", после того как код был вставлен наша база готова к работе с сервисом (Пользователь postgres и пароль в .env файле).

* Далее нужно установть WSL (Комнда для установки: "wsl --install") если его нет, установить любой дистрибутив linux (у меня Ubuntu), затем установить Redis в Ubuntu (Комнда для установки: "sudo apt install redis-server") и запустить его (Команда для запуска: "sudo service redis-server start"), для проверки того что Redis работает (Команда: "redis-cli ping"), если работает в ответ терминал выдаст "PONG"

* После того как были выполнены предыдущие шаги можно наконец запустить сервис, переходим по расположению файла main.go в консоли (куда вы его скачали) и запускаем (Команда для запуска: "go run main.go")

## Что нужно для запуска через Docker:


## Curl запросы:

1. Добавление пользователя и зачисление на счёт средств

Запрос: curl -X POST "http://localhost:8080/balance/add" ^ -H "Content-Type: application/json" ^ -d "{\"user_id\": 75, \"amount\": 250.00}"
   
Ответ: {"message":"Деньги добавлены на баланс"}

2. Добавление имени пользователю

Запрос: curl -X POST "http://localhost:8080/username/add" ^ -H "Content-Type: application/json" ^ -d "{\"user_id\": 75, \"name\": \"KirillDarkStalker98\"}"
   
Ответ: {"message":"Пользователь успешно добавлен"}

3. Получение баланса

Запрос: curl -X GET http://localhost:8080/balance/75
   
Ответ: {"balance":250,"user_id":75}

8. Добавление услуги в базу данных (curl -X POST http://localhost:8080/services/add -H "Content-Type: application/json" -d "{\"service_id\": 1001, \"service_name\": \"Service\"}"
curl -X POST http://localhost:8080/services/add -H "Content-Type: application/json" -d "{\"service_id\": 98, \"service_name\": \"DarkServise\"}"
{"message":"Service added successfully"} {"message":"Услуга успешно добавлена"})

9. Обновление названия услуги (curl -X POST http://localhost:8080/services/update -H "Content-Type: application/json" -d "{\"service_id\": 98, \"service_name\": \"DarkService98\"}"
{"message":"Услуга успешно обновлена"})

10. Удаление услуги (curl -X DELETE http://localhost:8080/services/delete -H "Content-Type: application/json" -d "{\"service_id\": 1423}"
{"message":"Услуга успешно удалена"})

11. Резервирование средств на отдельном счёте для покупки услуги (curl -X POST http://localhost:8080/funds/reserve -H "Content-Type: application/json" -d "{\"user_id\": 1, \"service_id\": 1001, \"order_id\": 1, \"amount\": 100.0}"
{"message":"Funds reserved successfully"}
curl -X POST http://localhost:8080/funds/reserve -H "Content-Type: application/json" -d "{\"user_id\": 75, \"service_id\": 98, \"order_id\": 98, \"amount\": 150.0}"
{"message":"Деньги для покупки услуги успешно зарезервированы (Дождитесь обработки покупки)"})

ГОВНО. (МБ) Признание выручки SAME (curl -X POST http://localhost:8080/funds/recognize -H "Content-Type: application/json" -d "{\"user_id\": 1, \"service_id\": 1001, \"order_id\": 1, \"amount\": 100.0}"
No reserved funds found for the given user, service, and order: sql: no rows in result set)

8. Списание средств SAME (ПОЛУЧШ) (curl -X POST "http://localhost:8080/funds/deduct" ^ -H "Content-Type: application/json" ^ -d "{\"user_id\": 1, \"service_id\": 1001, \"order_id\": 1, \"amount\": 100.0, \"success\": true}"
{"message":"Service applied successfully"}
curl -X POST "http://localhost:8080/funds/deduct" ^ -H "Content-Type: application/json" ^ -d "{\"user_id\": 75, \"service_id\": 98, \"order_id\": 98, \"amount\": 150.0, \"success\": true}"
{"message":"Услуга успешно приобретена"})

9. Отчёт (C:\Windows\System32> (curl -X GET http://localhost:8080/report/2024/09
{"report_url":"/reports/monthly_report_2024_09.csv"}
curl -X GET http://localhost:8080/report/2024/10
{"report_url":"/reports/месячный_отчёт_2024_10.csv"})

10. Просмотр транзакций (curl -X GET "http://localhost:8080/transactions?page=1&limit=10&sort_by=amount"(ТРАНЗАКЦИИ ВСЕХ ПОЛЬЗОВАТЕЛЕЙ)
[
  {
    "transaction_id": 1,
    "user_id": 1,
    "service_name": "Service",
    "amount": 100,
    "transaction_date": "2024-09-19T10:59:16.723177Z",
    "comment": "Revenue recognized for order 1"
  },
  {
    "transaction_id": 2,
    "user_id": 1,
    "service_name": "Service",
    "amount": 100,
    "transaction_date": "2024-09-19T13:04:05.955149Z",
    "comment": "1"
  },
  {
    "transaction_id": 3,
    "user_id": 2,
    "service_name": "",
    "amount": 100,
    "transaction_date": "2024-09-20T12:50:22.729912Z",
    "comment": "Transfer to user 1"
  },
  {
    "transaction_id": 4,
    "user_id": 1,
    "service_name": "",
    "amount": 100,
    "transaction_date": "2024-09-20T12:50:22.729912Z",
    "comment": "Transfer from user 2"
  },
  {
    "transaction_id": 5,
    "user_id": 2,
    "service_name": "",
    "amount": 100,
    "transaction_date": "2024-09-20T13:02:01.41351Z",
    "comment": "Account deposit"
  }
]
curl -X GET "http://localhost:8080/transactions?page=1&limit=10&sort_by=amount&user_id=1"(ТРАНЗАКЦИИ ВЫБРАННОГО ПОЛЬЗОВАТЕЛЯ)[
  {
    "transaction_id": 2,
    "user_id": 1,
    "service_name": "Service",
    "amount": 100,
    "transaction_date": "2024-09-19T13:04:05.955149Z",
    "comment": "1"
  },
  {
    "transaction_id": 1,
    "user_id": 1,
    "service_name": "Service",
    "amount": 100,
    "transaction_date": "2024-09-19T10:59:16.723177Z",
    "comment": "Revenue recognized for order 1"
  },
  {
    "transaction_id": 4,
    "user_id": 1,
    "service_name": "",
    "amount": 100,
    "transaction_date": "2024-09-20T12:50:22.729912Z",
    "comment": "Transfer from user 2"
  }
]
curl -X GET "http://localhost:8080/transactions?page=1&limit=10&sort_by=amount&user_id=75"
[
  {
    "transaction_id": 25,
    "user_id": 75,
    "service_name": "",
    "amount": 250,
    "transaction_date": "2024-10-03T13:07:56.829947Z",
    "comment": "Пополнение счета"
  },
  {
    "transaction_id": 26,
    "user_id": 75,
    "service_name": "DarkService98",
    "amount": 150,
    "transaction_date": "2024-10-03T13:19:04.28019Z",
    "comment": "Выручка признана по заказу 98"
  },
  {
    "transaction_id": 31,
    "user_id": 75,
    "service_name": "",
    "amount": 50,
    "transaction_date": "2024-10-03T13:29:02.556594Z",
    "comment": "Перевод пользователю 25"
  })

11. Перевод денег (curl -X POST http://localhost:8080/funds/transfer -H "Content-Type: application/json" -d "{\"from_user_id\":1,\"to_user_id\":2,\"amount\":100.00}"
curl -X POST http://localhost:8080/funds/transfer -H "Content-Type: application/json" -d "{\"from_user_id\":75,\"to_user_id\":25,\"amount\":50.00}"
{"message":"Средства переведены успешно"}
curl -X POST http://localhost:8080/funds/transfer -H "Content-Type: application/json" -d "{\"from_user_id\":75,\"to_user_id\":25,\"amount\":50.00}"
{"message":"Средства переведены успешно"})

## Файлы сервиса и их назначение:

go.mod - Нужен для запуска на windows

go.sum - Нужен для запуска на windows

SERVICE/ - Главная папка со всеми файлами

├── go.mod - Нужен для загрузки в docker

├── go.sum - Нужен для загрузки в docker

├── main.go - Файл для запуска микросервиса

├── bd.sql - Файл со скриптом для базы данных

├── docker-compose.yml - Файл для загрузки в Docker

├── Dockerfile - Файл для загрузки в Docker

├── service/ - Папка со всеми модулями и кодом

│   ├── bd.sql

│   ├── handlers.go - Файл со всеми функциями сервиса

│   ├── service.go - Панель управлениями функциями (Принимает запросы к методам)

│   ├── service_test.go - Файл с тестами кода

│   ├── db.go - Файл который подключает базу

│   ├── DataBase.env - Файл с данными к какой базе данных подключаться

│   └── Swagger.yaml - Swagger файл для API

