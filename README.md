# Microservice-for-working-with-user-balance (Микросервис для работы с балансом пользователей)

Функционал API

* Создание пользователя и добавление ему имени

* Зачисление средств

* Получение баланса пользователя 

* Покупка услуг и списание средств (зачисление средств на резервный счёт перед покупкой и списание средств с него)

* Перевод средств от пользователя к пользователю 

* Создание, изменение и удаление услуг которые можно купить

* Получение месячного отчёта 

* Просмотр транзакций всех пользователей или одного пользователя. 

Представляет из себя HTTP API с форматом JSON как при отправке запроса, так и при получении результата.

Файлы сервиса и их назначение:

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

* Что нужно для запуска на windows:

1. PostgreSQL
							
2. Язык Golang
	
3. Система очередей Redis (запуск через WSL)

* В начале для работы с сервисом в PostgreSQL необходимо создать базу данных с названием "service_db", затем выбрать базу, которая была создана, и прямо в консоль вставить SQL код из файла "bd.sql", после того как код был вставлен наша база готова к работе с сервисом (Пользователь postgres и пароль в .env файле).

* Далее нужно установть WSL (Комнда для установки: "wsl --install") если его нет, установить любой дистрибутив linux (у меня Ubuntu), затем установить Redis в Ubuntu (Комнда для установки: "sudo apt install redis-server") и запустить его (Команда для запуска: "sudo service redis-server start"), для проверки того что Redis работает (Команда: "redis-cli ping"), если работает в ответ терминал выдаст "PONG"

* После того как были выполнены предыдущие шаги можно наконец запустить сервис, переходим по расположению файла main.go в консоли (куда вы его скачали) и запускаем (Команда для запуска: "go run main.go")

Что нужно для запуска через Docker:

