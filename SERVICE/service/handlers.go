package SERVICE

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"

	_ "github.com/lib/pq"
	"golang.org/x/text/encoding/charmap"
)

var rdb *redis.Client
var ctx = context.Background()

// Асинхронный метод для добавления баланса с использованием системы очередей redis
func addBalance(w http.ResponseWriter, r *http.Request) {
	type Balance struct {
		UserID int     `json:"user_id"`
		Amount float64 `json:"amount"`
	}
	var balance Balance
	err := json.NewDecoder(r.Body).Decode(&balance)

	if err != nil {
		http.Error(w, "Неверный ввод: "+err.Error(), http.StatusBadRequest)
		return
	}

	if balance.Amount <= 0 {
		http.Error(w, "Сумма должна быть положительной", http.StatusBadRequest)
		return
	}

	// Проверка, существует ли пользователь с данным user_id
	var exists bool
	err = bd.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_id=$1)", balance.UserID).Scan(&exists)
	if err != nil {
		http.Error(w, "Не удалось проверить существование пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Если пользователя нет
	if !exists {
		_, err = bd.Exec("INSERT INTO users (user_id) VALUES ($1)", balance.UserID)
		if err != nil {
			http.Error(w, "Не удалось создать пользователя: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Формирование задачи для отправки в очередь Redis
	task := map[string]interface{}{
		"user_id": balance.UserID,
		"amount":  balance.Amount,
	}
	taskData, err := json.Marshal(task)
	if err != nil {
		http.Error(w, "Не удалось сериализовать: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Публикация задачи в канал Redis
	err = rdb.Publish(ctx, "balance_queue", taskData).Err()
	if err != nil {
		http.Error(w, "Не удалось опубликовать задачу в Redis: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Деньги добавлены на баланс"})
}

// Функция для обработки задач из очереди Redis на пополнение баланса
func processQueueBalance() {
	subscriber := rdb.Subscribe(ctx, "balance_queue")
	ch := subscriber.Channel()

	for msg := range ch {
		// Обрабатка каждого сообщения
		var task map[string]interface{}
		err := json.Unmarshal([]byte(msg.Payload), &task)
		if err != nil {
			log.Println("Не удалось десериализовать задачу: ", err)
			continue
		}

		userID := int(task["user_id"].(float64))
		amount := task["amount"].(float64)

		// Обновление баланса
		query := `
			INSERT INTO balances (user_id, balance) 
			VALUES ($1, $2) 
			ON CONFLICT (user_id) DO UPDATE SET balance = balances.balance + $2
		`
		_, err = bd.Exec(query, userID, amount)
		if err != nil {
			log.Println("Не удалось обновить баланс пользователя", userID, ":", err)
			continue
		}

		// Запись транзакции
		_, err = bd.Exec(
			"INSERT INTO transactions (user_id, service_id, amount, transaction_type, comment) VALUES ($1, NULL, $2, 'Пополнение', 'Пополнение счета')",
			userID, amount,
		)
		if err != nil {
			log.Println("Не удалось записать транзакцию для пользователя", userID, ":", err)
			continue
		}

		log.Println("Обработано обновление баланса для пользователя:", userID)
	}
}

// Метод добавления имени пользователю
func addUserName(w http.ResponseWriter, r *http.Request) {
	var user struct {
		UserID int    `json:"user_id"`
		Name   string `json:"name"`
	}

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Неверный ввод: "+err.Error(), http.StatusBadRequest)
		return
	}

	if user.UserID == 0 || user.Name == "" {
		http.Error(w, "Необходимо указать user_id и name", http.StatusBadRequest)
		return
	}

	// Проверка, существует ли пользователь с данным user_id
	var exists bool
	err = bd.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_id=$1)", user.UserID).Scan(&exists)
	if err != nil {
		http.Error(w, "Ошибка при проверке существования пользователя: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Если пользователя нет, создание его с указанным именем
	if !exists {
		_, err = bd.Exec("INSERT INTO users (user_id, name) VALUES ($1, $2)", user.UserID, user.Name)
		if err != nil {
			http.Error(w, "Ошибка при создании пользователя: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Если пользователь уже существует, обновление его имени
		_, err = bd.Exec("UPDATE users SET name=$1 WHERE user_id=$2", user.Name, user.UserID)
		if err != nil {
			http.Error(w, "Ошибка при обновлении имени пользователя: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Пользователь успешно добавлен"})
}

// Асинхронный метод получения баланса
func getBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["user_id"])

	if err != nil {
		http.Error(w, "Неверный идентификатор пользователя (user_id)", http.StatusBadRequest)
		return
	}

	//Канал для передачи результата
	resultChan := make(chan struct {
		balance float64
		err     error
	})

	// Запуск горутины для выполнения запроса к БД
	go func() {
		var balance float64
		err := bd.QueryRow("SELECT balance FROM balances WHERE user_id = $1", userID).Scan(&balance)
		resultChan <- struct {
			balance float64
			err     error
		}{balance: balance, err: err}
	}()

	// Результат
	result := <-resultChan

	if result.err == sql.ErrNoRows {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	} else if result.err != nil {
		http.Error(w, result.err.Error(), http.StatusInternalServerError)
		return
	}

	// Успешный ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": userID,
		"balance": result.balance,
	})
}

// Метод добавления услуги
func addService(w http.ResponseWriter, r *http.Request) {
	var newService struct {
		ServiceID   int    `json:"service_id,omitempty"`
		ServiceName string `json:"service_name"`
	}

	err := json.NewDecoder(r.Body).Decode(&newService)

	if err != nil {
		http.Error(w, "Неверный ввод: "+err.Error(), http.StatusBadRequest)
		return
	}

	if newService.ServiceName == "" {
		http.Error(w, "Укажите название услуги", http.StatusBadRequest)
		return
	}

	// Добавление услуги в БД
	var query string
	var args []interface{}

	if newService.ServiceID > 0 {
		// Если service_id указан
		query = "INSERT INTO services (service_id, service_name) VALUES ($1, $2)"
		args = []interface{}{newService.ServiceID, newService.ServiceName}
	} else {
		// Если service_id не указан
		query = "INSERT INTO services (service_name) VALUES ($1)"
		args = []interface{}{newService.ServiceName}
	}

	_, err = bd.Exec(query, args...)
	if err != nil {
		http.Error(w, "Не удалось добавить услугу: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Успешный ответ
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Услуга успешно добавлена"})
}

// Метод обновления или изменения названия услуги
func updateService(w http.ResponseWriter, r *http.Request) {
	var updatedService struct {
		ServiceID   int    `json:"service_id"`
		ServiceName string `json:"service_name"`
	}

	err := json.NewDecoder(r.Body).Decode(&updatedService)

	if err != nil {
		http.Error(w, "Неверный ввод: "+err.Error(), http.StatusBadRequest)
		return
	}

	if updatedService.ServiceID == 0 || updatedService.ServiceName == "" {
		http.Error(w, "Требуется идентификатор и название услуги", http.StatusBadRequest)
		return
	}

	// Обновление названия услуги в БД
	query := "UPDATE services SET service_name=$1 WHERE service_id=$2"
	_, err = bd.Exec(query, updatedService.ServiceName, updatedService.ServiceID)
	if err != nil {
		http.Error(w, "Не удалось обновить название услуги: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Услуга успешно обновлена"})
}

// Метод удаления услуги
func deleteService(w http.ResponseWriter, r *http.Request) {
	var serviceToDelete struct {
		ServiceID int `json:"service_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&serviceToDelete)

	if err != nil {
		http.Error(w, "Неверный ввод: "+err.Error(), http.StatusBadRequest)
		return
	}

	if serviceToDelete.ServiceID == 0 {
		http.Error(w, "Требуется идентификатор услуги", http.StatusBadRequest)
		return
	}

	// Удаление услуги из БД
	query := "DELETE FROM services WHERE service_id=$1"
	_, err = bd.Exec(query, serviceToDelete.ServiceID)
	if err != nil {
		http.Error(w, "Не удалось удалить услугу: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Услуга успешно удалена"})
}

// Асинхронный метод резервирования средств с использованием системы очередей redis
func reserveFunds(w http.ResponseWriter, r *http.Request) {
	var reservation struct {
		UserID    int     `json:"user_id"`
		ServiceID int     `json:"service_id"`
		OrderID   int     `json:"order_id"`
		Amount    float64 `json:"amount"`
	}

	err := json.NewDecoder(r.Body).Decode(&reservation)

	if err != nil {
		http.Error(w, "Неверный ввод: "+err.Error(), http.StatusBadRequest)
		return
	}

	if reservation.Amount <= 0 {
		http.Error(w, "Сумма должна быть положительной", http.StatusBadRequest)
		return
	}

	reservationJSON, err := json.Marshal(reservation)
	if err != nil {
		http.Error(w, "Не удалось сериализовать: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Добавление задачи в очередь Redis
	err = rdb.LPush(ctx, "reservation_queue", reservationJSON).Err()
	if err != nil {
		http.Error(w, "Не удалось добавить задачу в очередь: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Успешный ответ
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"message": "Деньги для покупки услуги успешно зарезервированы (Дождитесь обработки покупки)"})
}

// Функция для обработки задач из очереди Redis на резервирование средств
func processQueueReserveFunds() {
	for {
		// Извлечение задачи из очереди
		val, err := rdb.BRPop(ctx, 0, "reservation_queue").Result()
		if err != nil {
			fmt.Println("Ошибка при извлечении из очереди:", err)
			continue
		}
		var reservation struct {
			UserID    int     `json:"user_id"`
			ServiceID int     `json:"service_id"`
			OrderID   int     `json:"order_id"`
			Amount    float64 `json:"amount"`
		}
		err = json.Unmarshal([]byte(val[1]), &reservation)
		if err != nil {
			fmt.Println("Не удалось десериализовать задачу:", err)
			continue
		}

		// Проверка, достаточно ли средств на балансе
		var currentBalance float64
		err = bd.QueryRow("SELECT balance FROM balances WHERE user_id=$1", reservation.UserID).Scan(&currentBalance)
		if err != nil {
			fmt.Println("Пользователь не найден:", err)
			continue
		}

		if currentBalance < reservation.Amount {
			fmt.Println("На балансе недостаточно средств:", reservation.UserID)
			continue
		}
		// Запуск транзакции
		tx, err := bd.Begin()
		if err != nil {
			fmt.Println("Не удалось начать транзакцию:", err)
			continue
		}

		// Снятие средств с основного баланса
		_, err = tx.Exec("UPDATE balances SET balance = balance - $1 WHERE user_id = $2", reservation.Amount, reservation.UserID)
		if err != nil {
			tx.Rollback()
			fmt.Println("Не удалось снять средства с баланса:", err)
			continue
		}

		// Резервирование средств
		_, err = tx.Exec("INSERT INTO reserved_funds (user_id, service_id, order_id, amount) VALUES ($1, $2, $3, $4)",
			reservation.UserID, reservation.ServiceID, reservation.OrderID, reservation.Amount)
		if err != nil {
			tx.Rollback()
			fmt.Println("Не удалось зарезервировать средства:", err)
			continue
		}

		// Подтверждение транзакции
		err = tx.Commit()
		if err != nil {
			fmt.Println("Не удалось подтвердить транзакцию:", err)
			continue
		}

		fmt.Printf("Средства успешно зарезервированы для пользователя %d\n", reservation.UserID)
	}
}

// Реализация метода списания средств
func deductFunds(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		UserID    int     `json:"user_id"`
		ServiceID int     `json:"service_id"`
		OrderID   int     `json:"order_id"`
		Amount    float64 `json:"amount"`
		Success   bool    `json:"success"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)

	if err != nil {
		http.Error(w, "Неверный ввод: "+err.Error(), http.StatusBadRequest)
		return
	}

	if requestData.Amount <= 0 {
		http.Error(w, "Сумма должна быть положительной", http.StatusBadRequest)
		return
	}

	// Запуск транзакции
	tx, err := bd.Begin()
	if err != nil {
		http.Error(w, "Не удалось начать транзакцию: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Проверка, есть ли средства в резерве
	var reservedAmount float64
	err = tx.QueryRow(
		"SELECT amount FROM reserved_funds WHERE user_id=$1 AND service_id=$2 AND order_id=$3",
		requestData.UserID, requestData.ServiceID, requestData.OrderID,
	).Scan(&reservedAmount)

	if err != nil {
		tx.Rollback()
		http.Error(w, "Зарезервированные средства не найдены: "+err.Error(), http.StatusBadRequest)
		return
	}

	if reservedAmount < requestData.Amount {
		tx.Rollback()
		http.Error(w, "Недостаточно зарезервированных средств", http.StatusBadRequest)
		return
	}

	if requestData.Success {
		// Если услуга применена успешно, списание средств из резерва
		_, err = tx.Exec(
			"DELETE FROM reserved_funds WHERE user_id=$1 AND service_id=$2 AND order_id=$3",
			requestData.UserID, requestData.ServiceID, requestData.OrderID,
		)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Не удалось списать зарезервированные средства: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Запись транзакции в таблицу
		comment := fmt.Sprintf("Выручка признана по заказу %d", requestData.OrderID)
		_, err = tx.Exec(
			"INSERT INTO transactions (user_id, service_id, amount, transaction_type, comment) VALUES ($1, $2, $3, 'Покупка', $4)",
			requestData.UserID, requestData.ServiceID, requestData.Amount, comment,
		)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Не удалось записать транзакцию: "+err.Error(), http.StatusInternalServerError)
			return
		}

	} else {
		// Если услуга не применена, возврат средств на баланс
		_, err = tx.Exec(
			"UPDATE balances SET balance = balance + $1 WHERE user_id = $2",
			requestData.Amount, requestData.UserID,
		)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Не удалось вернуть зарезервированные средства: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Удаление резерва после возврата средств
		_, err = tx.Exec(
			"DELETE FROM reserved_funds WHERE user_id=$1 AND service_id=$2 AND order_id=$3",
			requestData.UserID, requestData.ServiceID, requestData.OrderID,
		)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Не удалось удалить зарезервированные средства после возврата: "+err.Error(), http.StatusInternalServerError)
			return
		}
		comment2 := fmt.Sprintf("Возврат денег за не оказанную услугу по заказу %d", requestData.OrderID)

		// Запись транзакции возврата
		_, err = tx.Exec(
			"INSERT INTO transactions (user_id, service_id, amount, transaction_type, comment) VALUES ($1, $2, $3, 'Возврат', $4)",
			requestData.UserID, requestData.ServiceID, requestData.Amount, comment2,
		)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Не удалось записать транзакцию возврата: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Подтверждение транзакции
	err = tx.Commit()
	if err != nil {
		http.Error(w, "Не удалось подтвердить транзакцию: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Успешный ответ
	w.WriteHeader(http.StatusOK)
	if requestData.Success {
		json.NewEncoder(w).Encode(map[string]string{"message": "Услуга успешно приобретена"})
	} else {
		json.NewEncoder(w).Encode(map[string]string{"message": "Возврат средств из-за не оказанной услуги"})
	}
}

// Асинхронный метод перевода средств
func transferFunds(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		FromUserID int     `json:"from_user_id"`
		ToUserID   int     `json:"to_user_id"`
		Amount     float64 `json:"amount"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)

	if err != nil {
		http.Error(w, "Неверный ввод: "+err.Error(), http.StatusBadRequest)
		return
	}

	if requestData.Amount <= 0 {
		http.Error(w, "Сумма должна быть положительной", http.StatusBadRequest)
		return
	}

	taskData, err := json.Marshal(requestData)
	if err != nil {
		http.Error(w, "Не удалось сериализовать: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Добавление задачи в Redis очередь
	err = rdb.RPush(r.Context(), "funds_transfer_queue", taskData).Err()
	if err != nil {
		http.Error(w, "Не удалось добавить задачу в очередь: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Успешный ответ
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"message": "Средства переведены успешно"})
}

// Функция для обработки задач из Redis перевода средств
func processQueueFundsTransfers() {
	for {
		// Извлечение задачи из очереди
		taskData, err := rdb.LPop(context.Background(), "funds_transfer_queue").Result()
		if err == redis.Nil {
			time.Sleep(1 * time.Second)
			continue
		} else if err != nil {
			log.Printf("Ошибка получения задачи: %v\n", err)
			continue
		}

		var requestData struct {
			FromUserID int     `json:"from_user_id"`
			ToUserID   int     `json:"to_user_id"`
			Amount     float64 `json:"amount"`
		}
		err = json.Unmarshal([]byte(taskData), &requestData)
		if err != nil {
			log.Printf("Не удалось десериализовать данные: %v\n", err)
			continue
		}

		// Перевод средств
		err = transferFundsAsync(requestData.FromUserID, requestData.ToUserID, requestData.Amount)
		if err != nil {
			log.Printf("Не удалось перевести средства: %v\n", err)
			continue
		}

		log.Printf("Средства успешно переведены от пользователя %d пользователю %d\n", requestData.FromUserID, requestData.ToUserID)
	}

}

// Метод для обработки перевода средств
func transferFundsAsync(fromUserID, toUserID int, amount float64) error {
	// Запуск транзакции
	tx, err := bd.Begin()
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %v", err)
	}

	// Проверка баланса отправителя
	var senderBalance float64
	err = tx.QueryRow(
		"SELECT balance FROM balances WHERE user_id=$1",
		fromUserID,
	).Scan(&senderBalance)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("отправитель не найден: %v", err)
	}

	if senderBalance < amount {
		tx.Rollback()
		return fmt.Errorf("недостаточно средств")
	}

	// Обновление баланса отправителя (-)
	_, err = tx.Exec(
		"UPDATE balances SET balance = balance - $1 WHERE user_id = $2",
		amount, fromUserID,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("не удалось обновить баланс отправителя: %v", err)
	}

	// Обновление баланса получателя (+)
	_, err = tx.Exec(
		"UPDATE balances SET balance = balance + $1 WHERE user_id = $2",
		amount, toUserID,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("не удалось обновить баланс получателя: %v", err)
	}

	// Запись транзакции
	_, err = tx.Exec(
		"INSERT INTO transactions (user_id, service_id, amount, transaction_type, comment) VALUES ($1, NULL, $2, 'Перевод', $3)",
		fromUserID, amount, fmt.Sprintf("Перевод пользователю %d", toUserID),
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("не удалось записать транзакцию отправителя: %v", err)
	}

	_, err = tx.Exec(
		"INSERT INTO transactions (user_id, service_id, amount, transaction_type, comment) VALUES ($1, NULL, $2, 'Перевод', $3)",
		toUserID, amount, fmt.Sprintf("Перевод от пользователя %d", fromUserID),
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("не удалось записать транзакцию получателя: %v", err)
	}

	// Подтверждение транзакции
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("не удалось подтвердить транзакцию: %v", err)
	}

	return nil
}

// Асинхронный метод получения месячного отчёта
func getMonthlyReport(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	year, err := strconv.Atoi(vars["year"])
	if err != nil {
		http.Error(w, "Неверный формат года", http.StatusBadRequest)
		return
	}

	month, err := strconv.Atoi(vars["month"])
	if err != nil {
		http.Error(w, "Неверный формат месяца", http.StatusBadRequest)
		return
	}

	if month < 1 || month > 12 {
		http.Error(w, "Месяц должен быть в диапазоне от 1 до 12", http.StatusBadRequest)
		return
	}

	fileNameCh := make(chan string)
	errCh := make(chan error)

	go func() {
		fileName, err := generateMonthlyReport(ctx, year, month)
		if err != nil {
			errCh <- err
			return
		}
		fileNameCh <- fileName
	}()

	select {
	case <-ctx.Done():
		http.Error(w, "Время запроса истекло", http.StatusGatewayTimeout)
	case err := <-errCh:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	case fileName := <-fileNameCh:
		w.Header().Set("Content-Type", "application/json")
		// Успешный ответ
		json.NewEncoder(w).Encode(map[string]string{"report_url": "/reports/" + fileName})
	}
}

// Функция для генерации отчёта
func generateMonthlyReport(ctx context.Context, year, month int) (string, error) {
	if month < 1 || month > 12 {
		return "", fmt.Errorf("неверный месяц: %d", month)
	}
	query := `
    SELECT services.service_name, SUM(transactions.amount) as total_revenue
    FROM transactions
    JOIN services ON services.service_id = transactions.service_id
    WHERE EXTRACT(YEAR FROM transaction_date) = $1 AND EXTRACT(MONTH FROM transaction_date) = $2
    GROUP BY services.service_name;
    `

	rowsCh := make(chan *sql.Rows)
	errCh := make(chan error)

	go func() {
		rows, err := bd.QueryContext(ctx, query, year, month)
		if err != nil {
			errCh <- err
			return
		}
		rowsCh <- rows
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case err := <-errCh:
		return "", err
	case rows := <-rowsCh:
		defer rows.Close()

		fileName := fmt.Sprintf("месячный_отчёт_%d_%02d.csv", year, month)
		file, err := os.Create(fileName)
		if err != nil {
			return "", err
		}
		defer file.Close()

		// Использование charmap.Windows1251 для записи русских символов в правильной кодировке
		writer := csv.NewWriter(charmap.Windows1251.NewEncoder().Writer(file))
		writer.Comma = ';'
		writer.Write([]string{"Название услуги", "Доход за месяц"})

		for rows.Next() {
			var serviceName string
			var totalRevenue float64
			rows.Scan(&serviceName, &totalRevenue)
			writer.Write([]string{serviceName, fmt.Sprintf("%.2f", totalRevenue)})
		}

		writer.Flush()
		return fileName, nil
	}
}

// Асинхронный метод получения транзакций
func getTransactions(w http.ResponseWriter, r *http.Request) {
	type Transaction struct {
		TransactionID   int     `json:"transaction_id"`
		UserID          int     `json:"user_id"`
		ServiceName     string  `json:"service_name"`
		Amount          float64 `json:"amount"`
		TransactionDate string  `json:"transaction_date"`
		Comment         string  `json:"comment"`
	}
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page <= 0 {
		http.Error(w, "Неверный ввод: параметр 'page' должен быть положительным целым числом", http.StatusBadRequest)
		return
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit <= 0 {
		http.Error(w, "Неверный ввод: параметр 'limit' должен быть положительным целым числом", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(r.URL.Query().Get("user_id"))
	if err != nil && r.URL.Query().Get("user_id") != "" {
		http.Error(w, "Неверный ввод: параметр 'user_id' должен быть целым числом", http.StatusBadRequest)
		return
	}

	sortBy := r.URL.Query().Get("sort_by")
	if sortBy != "" && sortBy != "amount" && sortBy != "transaction_date" {
		http.Error(w, "Неверный ввод: параметр 'sort_by' может быть либо 'amount', либо 'transaction_date'", http.StatusBadRequest)
		return
	}

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	offset := (page - 1) * limit
	sortQuery := "ORDER BY transaction_date DESC"
	if sortBy == "amount" {
		sortQuery = "ORDER BY amount DESC"
	}

	query := `
    SELECT t.transaction_id, t.user_id, COALESCE(s.service_name, '') AS service_name, t.amount, t.transaction_date, t.comment
    FROM transactions t
    LEFT JOIN services s ON s.service_id = t.service_id
    `
	if userID > 0 {
		query += "WHERE t.user_id = $3 "
	}
	query += sortQuery + " LIMIT $1 OFFSET $2"

	// Каналы для передачи результатов
	rowsCh := make(chan *sql.Rows)
	errCh := make(chan error)

	// Выполнение SQL-запроса в горутине
	go func() {
		var rows *sql.Rows
		var err error
		if userID > 0 {
			rows, err = bd.Query(query, limit, offset, userID)
		} else {
			rows, err = bd.Query(query, limit, offset)
		}

		if err != nil {
			errCh <- err
			return
		}
		rowsCh <- rows
	}()

	select {
	case err := <-errCh:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	case rows := <-rowsCh:
		defer rows.Close()

		var transactions []Transaction
		for rows.Next() {
			var trans Transaction
			err := rows.Scan(&trans.TransactionID, &trans.UserID, &trans.ServiceName, &trans.Amount, &trans.TransactionDate, &trans.Comment)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			transactions = append(transactions, trans)
		}

		w.Header().Set("Content-Type", "application/json")

		jsonResponse, err := json.MarshalIndent(transactions, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Успешный ответ
		w.Write(jsonResponse)
	}
}
