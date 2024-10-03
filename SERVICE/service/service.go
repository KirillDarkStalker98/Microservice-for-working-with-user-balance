package SERVICE

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func SERVICE() {

	initDB()
	defer bd.Close()
	go processQueueBalance()
	go processQueueReserveFunds()
	go processQueueFundsTransfers()

	router := mux.NewRouter()

	// Обработчики для управления пользователем и балансом
	router.HandleFunc("/username/add", addUserName).Methods("POST")
	router.HandleFunc("/balance/add", addBalance).Methods("POST")
	router.HandleFunc("/balance/{user_id}", getBalance).Methods("GET")

	// Обработчики для управления услугами
	router.HandleFunc("/services/add", addService).Methods("POST")
	router.HandleFunc("/services/update", updateService).Methods("POST")
	router.HandleFunc("/services/delete", deleteService).Methods("DELETE")

	// Обработчик для резервирования средств
	router.HandleFunc("/funds/reserve", reserveFunds).Methods("POST")

	// Обработчик для признания выручки (списание из резерва)
	router.HandleFunc("/funds/deduct", deductFunds).Methods("POST")

	// Обработчик для перевода средств между пользователями
	router.HandleFunc("/funds/transfer", transferFunds).Methods("POST")

	// Обработчик для получения месячного отчета
	router.HandleFunc("/report/{year}/{month}", getMonthlyReport).Methods("GET")

	// Обработчик для получения списка транзакций с пагинацией и сортировкой
	router.HandleFunc("/transactions", getTransactions).Methods("GET")

	// Запуск сервера
	log.Println("Сервер работает на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
