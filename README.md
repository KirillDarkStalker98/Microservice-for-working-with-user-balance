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

### 1. Добавление пользователя и зачисление на счёт средств

* Запрос: curl -X POST "http://localhost:8080/balance/add" ^ -H "Content-Type: application/json" ^ -d "{\"user_id\": 75, \"amount\": 250.00}"
   
* Ответ: {"message":"Деньги добавлены на баланс"}

### 2. Добавление имени пользователю

* Запрос: curl -X POST "http://localhost:8080/username/add" ^ -H "Content-Type: application/json" ^ -d "{\"user_id\": 75, \"name\": \"KirillDarkStalker98\"}"
   
* Ответ: {"message":"Пользователь успешно добавлен"}

### 3. Получение баланса

* Запрос: curl -X GET http://localhost:8080/balance/75
   
* Ответ: {"balance":250,"user_id":75}

### 4. Добавление услуги в базу данных 

* Запрос: curl -X POST http://localhost:8080/services/add -H "Content-Type: application/json" -d "{\"service_id\": 98, \"service_name\": \"DarkServise\"}"
  
* Ответ: {"message":"Услуга успешно добавлена"}

### 5. Обновление названия услуги

* Запрос: curl -X POST http://localhost:8080/services/update -H "Content-Type: application/json" -d "{\"service_id\": 98, \"service_name\": \"DarkService98\"}"

* Ответ: {"message":"Услуга успешно обновлена"}

### 6. Удаление услуги 

* Запрос: curl -X DELETE http://localhost:8080/services/delete -H "Content-Type: application/json" -d "{\"service_id\": 98}"

* Ответ: {"message":"Услуга успешно удалена"}

### 7. Резервирование средств на отдельном счёте для покупки услуги 

* Запрос: curl -X POST http://localhost:8080/funds/reserve -H "Content-Type: application/json" -d "{\"user_id\": 75, \"service_id\": 98, \"order_id\": 98, \"amount\": 150.0}"

* Ответ: {"message":"Деньги для покупки услуги успешно зарезервированы (Дождитесь обработки покупки)"}

### 8. Списание средств 

* Запрос: curl -X POST "http://localhost:8080/funds/deduct" ^ -H "Content-Type: application/json" ^ -d "{\"user_id\": 75, \"service_id\": 98, \"order_id\": 98, \"amount\": 150.0, \"success\": true}"

* Ответ: {"message":"Услуга успешно приобретена"}

### 9. Отчёт 

* Запрос: curl -X GET http://localhost:8080/report/2024/10
  
* Ответ: {"report_url":"/reports/месячный_отчёт_2024_10.csv"}

### 10. Просмотр транзакций 
* Запрос: curl -X GET "http://localhost:8080/transactions?page=1&limit=10&sort_by=amount" (ТРАНЗАКЦИИ ВСЕХ ПОЛЬЗОВАТЕЛЕЙ)

* Запрос: curl -X GET "http://localhost:8080/transactions?page=1&limit=10&sort_by=amount&user_id=75" (ТРАНЗАКЦИИ ВЫБРАННОГО ПОЛЬЗОВАТЕЛЯ)
  
* Ответ:

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
  
  }
  

### 11. Перевод денег
    
* Запрос: curl -X POST http://localhost:8080/funds/transfer -H "Content-Type: application/json" -d "{\"from_user_id\":75,\"to_user_id\":25,\"amount\":50.00}"

* Ответ: {"message":"Средства переведены успешно"}

## Файлы сервиса и их назначение:

go.mod - Нужен для запуска на windows

go.sum - Нужен для запуска на windows

SERVICE/ - Главная папка со всеми файлами

├── go.mod - Нужен для загрузки в docker

├── go.sum - Нужен для загрузки в docker

├── main.go - Файл для запуска микросервиса

├── bd.sql - Файл со скриптами создания таблиц для докера

├── docker-compose.yml - Файл для загрузки в Docker

├── Dockerfile - Файл для загрузки в Docker

├── service/ - Папка со всеми модулями и кодом

│   ├── bd.sql - Файл со скриптами создания таблиц

│   ├── handlers.go - Файл со всеми функциями сервиса

│   ├── service.go - Файл управления функциями (Принимает запросы к методам)

│   ├── service_test.go - Файл с тестами кода

│   ├── db.go - Файл который подключает базу и Redis

│   ├── DataBase.env - Файл с данными к какой базе данных подключаться

│   └── Swagger.yaml - Swagger файл для API

