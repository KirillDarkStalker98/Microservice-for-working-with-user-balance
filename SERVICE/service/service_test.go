package SERVICE

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"
)

// Тесты запускать при созданной базе данных в PostgreSQL, включенном сервере и redis

func TestAddBalance(t *testing.T) {
	// Команда curl
	cmd := exec.Command("curl", "-s", "-X", "POST", "http://localhost:8080/balance/add",
		"-H", "Content-Type: application/json",
		"-d", `{"user_id": 25, "amount": 100.50}`)

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось выполнить curl запрос: %v", err)
	}

	// Проверяем вывод curl
	expectedOutput := `{"message":"Деньги добавлены на баланс"}`

	if strings.TrimSpace(string(output)) != expectedOutput {
		t.Errorf("Неверный ответ: получено %v, ожидалось %v", string(output), expectedOutput)
	}

}

func TestAddBalanceEnter(t *testing.T) {
	// Команда curl
	cmd := exec.Command("curl", "-s", "-X", "POST", "http://localhost:8080/balance/add",
		"-H", "Content-Type: application/json",
		"-d", `{"user_id": 25, "amount": a}`)

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось выполнить curl запрос: %v", err)
	}

	// Проверяем вывод curl
	expectedOutput := `Неверный ввод: invalid character 'a' looking for beginning of value`

	if strings.TrimSpace(string(output)) != expectedOutput {
		t.Errorf("Неверный ответ: получено %v, ожидалось %v", string(output), expectedOutput)
	}

}

// Так как некоторые обработчики ошибок будут встречатся в коде, я протестирую их 1 раз
func TestAddBalanceSum(t *testing.T) {
	// Команда curl
	cmd := exec.Command("curl", "-s", "-X", "POST", "http://localhost:8080/balance/add",
		"-H", "Content-Type: application/json",
		"-d", `{"user_id": 25, "amount": -100.50}`)

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось выполнить curl запрос: %v", err)
	}

	// Проверяем вывод curl
	expectedOutput := `Сумма должна быть положительной`

	if strings.TrimSpace(string(output)) != expectedOutput {
		t.Errorf("Неверный ответ: получено %v, ожидалось %v", string(output), expectedOutput)
	}

}

// Тест обработчика из на сериализацию
func TestJSONMarshalErrorHandling(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Создаем объект с некорректными данными для сериализации
		task := make(chan int) // Канал не может быть сериализован в JSON
		_, err := json.Marshal(task)
		if err != nil {
			http.Error(w, "Не удалось сериализовать: "+err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Создаем запрос и записываем ответ
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Не удалось создать запрос: %v", err)
	}

	// Используем httptest для создания ResponseRecorder
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Проверяем код ответа
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Неверный статус-код: получено %v, ожидалось %v", status, http.StatusInternalServerError)
	}

	// Проверяем сообщение об ошибке
	expectedError := "Не удалось сериализовать: json: unsupported type: chan int\n"
	if rr.Body.String() != expectedError {
		t.Errorf("Неверное сообщение об ошибке: получено %v, ожидалось %v", rr.Body.String(), expectedError)
	}
}

func TestAddUserName(t *testing.T) {
	// Команда curl
	cmd := exec.Command("curl", "-s", "-X", "POST", "http://localhost:8080/username/add",
		"-H", "Content-Type: application/json",
		"-d", `{"user_id": 25, "name": "Кирилл" }`)

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось выполнить curl запрос: %v", err)
	}

	// Проверяем вывод curl
	expectedOutput := `{"message":"Пользователь успешно добавлен"}`

	if strings.TrimSpace(string(output)) != expectedOutput {
		t.Errorf("Неверный ответ: получено %v, ожидалось %v", string(output), expectedOutput)
	}

}

func TestAddUserNameErr(t *testing.T) {
	// Команда curl
	cmd := exec.Command("curl", "-s", "-X", "POST", "http://localhost:8080/username/add",
		"-H", "Content-Type: application/json",
		"-d", `{"user_id": 25, "name": "" }`)

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось выполнить curl запрос: %v", err)
	}

	// Проверяем вывод curl
	expectedOutput := `Необходимо указать user_id и name`

	if strings.TrimSpace(string(output)) != expectedOutput {
		t.Errorf("Неверный ответ: получено %v, ожидалось %v", string(output), expectedOutput)
	}

}

func TestGetBalance(t *testing.T) {
	// Команда curl
	cmd := exec.Command("curl", "-s", "-X", "GET", "http://localhost:8080/balance/25")

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось выполнить curl запрос: %v", err)
	}

	// Проверяем вывод curl

	expectedOutput := `{"balance":100.5,"user_id":25}`

	if strings.TrimSpace(string(output)) != expectedOutput {
		t.Errorf("Неверный ответ: получено %v, ожидалось %v", string(output), expectedOutput)
	}

}

func TestGetBalanceErr(t *testing.T) {
	// Команда curl
	cmd := exec.Command("curl", "-s", "-X", "GET", "http://localhost:8080/balance/ad")

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось выполнить curl запрос: %v", err)
	}

	// Проверяем вывод curl

	expectedOutput := `Неверный идентификатор пользователя (user_id)`

	if strings.TrimSpace(string(output)) != expectedOutput {
		t.Errorf("Неверный ответ: получено %v, ожидалось %v", string(output), expectedOutput)
	}

}

func TestGetBalanceErrUser(t *testing.T) {
	// Команда curl
	cmd := exec.Command("curl", "-s", "-X", "GET", "http://localhost:8080/balance/533")

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось выполнить curl запрос: %v", err)
	}

	// Проверяем вывод curl

	expectedOutput := `Пользователь не найден`

	if strings.TrimSpace(string(output)) != expectedOutput {
		t.Errorf("Неверный ответ: получено %v, ожидалось %v", string(output), expectedOutput)
	}

}

func TestAddService(t *testing.T) {
	// Команда curl
	cmd := exec.Command("curl", "-s", "-X", "POST", "http://localhost:8080/services/add",
		"-H", "Content-Type: application/json",
		"-d", `{"service_id": 1423, "service_name": "Услуга: для теста"}`)

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось выполнить curl запрос: %v", err)
	}

	// Проверяем вывод curl

	expectedOutput := `{"message":"Услуга успешно добавлена"}`

	if strings.TrimSpace(string(output)) != expectedOutput {
		t.Errorf("Неверный ответ: получено %v, ожидалось %v", string(output), expectedOutput)
	}

}

func TestAddServiceEnter(t *testing.T) {
	// Команда curl
	cmd := exec.Command("curl", "-s", "-X", "POST", "http://localhost:8080/services/add",
		"-H", "Content-Type: application/json",
		"-d", `{"service_id": 1423, "service_name": ""}`)

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось выполнить curl запрос: %v", err)
	}

	// Проверяем вывод curl
	expectedOutput := `Укажите название услуги`

	if strings.TrimSpace(string(output)) != expectedOutput {
		t.Errorf("Неверный ответ: получено %v, ожидалось %v", string(output), expectedOutput)
	}

}

func TestUpdateService(t *testing.T) {
	// Команда curl
	cmd := exec.Command("curl", "-s", "-X", "POST", "http://localhost:8080/services/update",
		"-H", "Content-Type: application/json",
		"-d", `{"service_id": 1423, "service_name": "Услуга: для теста переименованная"}`)

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось выполнить curl запрос: %v", err)
	}

	// Проверяем вывод curl

	expectedOutput := `{"message":"Услуга успешно обновлена"}`

	if strings.TrimSpace(string(output)) != expectedOutput {
		t.Errorf("Неверный ответ: получено %v, ожидалось %v", string(output), expectedOutput)
	}

}

func TestUpdateServiceErr(t *testing.T) {
	// Команда curl
	cmd := exec.Command("curl", "-s", "-X", "POST", "http://localhost:8080/services/update",
		"-H", "Content-Type: application/json",
		"-d", `{"service_id": 1423, "service_name": ""}`)

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось выполнить curl запрос: %v", err)
	}

	// Проверяем вывод curl

	expectedOutput := `Требуется идентификатор и название услуги`

	if strings.TrimSpace(string(output)) != expectedOutput {
		t.Errorf("Неверный ответ: получено %v, ожидалось %v", string(output), expectedOutput)
	}

}

func TestDeleteService(t *testing.T) {
	// Команда curl
	cmd := exec.Command("curl", "-s", "-X", "DELETE", "http://localhost:8080/services/delete",
		"-H", "Content-Type: application/json",
		"-d", `{"service_id": 1423}`)

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось выполнить curl запрос: %v", err)
	}

	// Проверяем вывод curl

	expectedOutput := `{"message":"Услуга успешно удалена"}`

	if strings.TrimSpace(string(output)) != expectedOutput {
		t.Errorf("Неверный ответ: получено %v, ожидалось %v", string(output), expectedOutput)
	}

}

func TestReserveFunds(t *testing.T) {
	// Команда curl
	cmd := exec.Command("curl", "-s", "-X", "POST", "http://localhost:8080/funds/reserve",
		"-H", "Content-Type: application/json",
		"-d", `{"user_id": 25, "service_id": 1423, "order_id": 15, "amount": 100.0}`)

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось выполнить curl запрос: %v", err)
	}

	// Проверяем вывод curl

	expectedOutput := `{"message":"Деньги для покупки услуги успешно зарезервированы (Дождитесь обработки покупки)"}`

	if strings.TrimSpace(string(output)) != expectedOutput {
		t.Errorf("Неверный ответ: получено %v, ожидалось %v", string(output), expectedOutput)
	}

}

func TestDeductFunds(t *testing.T) {
	// Команда curl
	cmd := exec.Command("curl", "-s", "-X", "POST", "http://localhost:8080/funds/deduct",
		"-H", "Content-Type: application/json",
		"-d", `{"user_id": 25, "service_id": 1423, "order_id": 15, "amount": 100.0, "success": true}`)

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось выполнить curl запрос: %v", err)
	}

	// Проверяем вывод curl

	expectedOutput := `{"message":"Услуга успешно приобретена"}`

	if strings.TrimSpace(string(output)) != expectedOutput {
		t.Errorf("Неверный ответ: получено %v, ожидалось %v", string(output), expectedOutput)
	}

}

func TestTransferFunds(t *testing.T) {
	// Команда curl
	cmd := exec.Command("curl", "-s", "-X", "POST", "http://localhost:8080/funds/transfer",
		"-H", "Content-Type: application/json",
		"-d", `{"from_user_id":25,"to_user_id":1,"amount":0.50}`)

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось выполнить curl запрос: %v", err)
	}

	// Проверяем вывод curl

	expectedOutput := `{"message":"Средства переведены успешно"}`

	if strings.TrimSpace(string(output)) != expectedOutput {
		t.Errorf("Неверный ответ: получено %v, ожидалось %v", string(output), expectedOutput)
	}

}

func TestGetMonthlyReport(t *testing.T) {
	// Команда curl
	cmd := exec.Command("curl", "-s", "-X", "GET", "http://localhost:8080/report/2024/09")

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось выполнить curl запрос: %v", err)
	}

	// Проверяем вывод curl

	expectedOutput := `{"report_url":"/reports/месячный_отчёт_2024_09.csv"}`

	if strings.TrimSpace(string(output)) != expectedOutput {
		t.Errorf("Неверный ответ: получено %v, ожидалось %v", string(output), expectedOutput)
	}

}

func TestGetMonthlyReportErr(t *testing.T) {
	// Команда curl
	cmd := exec.Command("curl", "-s", "-X", "GET", "http://localhost:8080/report/2024/15")

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось выполнить curl запрос: %v", err)
	}

	// Проверяем вывод curl

	expectedOutput := `Месяц должен быть в диапазоне от 1 до 12`

	if strings.TrimSpace(string(output)) != expectedOutput {
		t.Errorf("Неверный ответ: получено %v, ожидалось %v", string(output), expectedOutput)
	}

}

func TestGetMonthlyReportErr2(t *testing.T) {
	// Команда curl
	cmd := exec.Command("curl", "-s", "-X", "GET", "http://localhost:8080/report/df/15")

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось выполнить curl запрос: %v", err)
	}

	// Проверяем вывод curl

	expectedOutput := `Неверный формат года`

	if strings.TrimSpace(string(output)) != expectedOutput {
		t.Errorf("Неверный ответ: получено %v, ожидалось %v", string(output), expectedOutput)
	}

}

func TestGetMonthlyReportErr3(t *testing.T) {
	// Команда curl
	cmd := exec.Command("curl", "-s", "-X", "GET", "http://localhost:8080/report/2024/D")

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось выполнить curl запрос: %v", err)
	}

	// Проверяем вывод curl

	expectedOutput := `Неверный формат месяца`

	if strings.TrimSpace(string(output)) != expectedOutput {
		t.Errorf("Неверный ответ: получено %v, ожидалось %v", string(output), expectedOutput)
	}

}

func TestGetTransactionErr(t *testing.T) {
	// Команда curl
	cmd := exec.Command("curl", "-s", "-X", "GET", "http://localhost:8080/transactions?page=df&limit=10&sort_by=amount")

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось выполнить curl запрос: %v", err)
	}

	// Проверяем вывод curl

	expectedOutput := `Неверный ввод: параметр 'page' должен быть положительным целым числом`

	if strings.TrimSpace(string(output)) != expectedOutput {
		t.Errorf("Неверный ответ: получено %v, ожидалось %v", string(output), expectedOutput)
	}

}

func TestGetTransactionErr2(t *testing.T) {
	// Команда curl
	cmd := exec.Command("curl", "-s", "-X", "GET", "http://localhost:8080/transactions?page=1&limit=10&sort_by=a")

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось выполнить curl запрос: %v", err)
	}

	// Проверяем вывод curl

	expectedOutput := `Неверный ввод: параметр 'sort_by' может быть либо 'amount', либо 'transaction_date'`

	if strings.TrimSpace(string(output)) != expectedOutput {
		t.Errorf("Неверный ответ: получено %v, ожидалось %v", string(output), expectedOutput)
	}

}

func TestGetTransactionErr3(t *testing.T) {
	// Команда curl
	cmd := exec.Command("curl", "-s", "-X", "GET", "http://localhost:8080/transactions?page=1&limit=df&sort_by=amount")

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Не удалось выполнить curl запрос: %v", err)
	}

	// Проверяем вывод curl

	expectedOutput := `Неверный ввод: параметр 'limit' должен быть положительным целым числом`

	if strings.TrimSpace(string(output)) != expectedOutput {
		t.Errorf("Неверный ответ: получено %v, ожидалось %v", string(output), expectedOutput)
	}

}
