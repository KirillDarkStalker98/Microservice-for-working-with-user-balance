package SERVICE

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var bd *sql.DB

func initDB() {
	env := "service/DataBase.env" //E:\\С диска о\\GO\\SERVICE\\service\\"DataBase.env" Пример пути до .env файла
	err := godotenv.Load(env)     //".env"
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	//Для Windows
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
	//Для докера
	/*connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
	os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))*/

	var errDb error
	bd, errDb = sql.Open("postgres", connStr)
	if errDb != nil {
		log.Fatal("Не удалось подключиться к базе данных: ", errDb)
	}

	errPing := bd.Ping()
	if errPing != nil {
		log.Fatal("Невозможно связаться с базой данных: ", errPing)
	}

	fmt.Println("База данных успешно подключена!")

	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // адрес Redis для windows
		//Addr: "redis:6379", // адрес Redis для докера
	})

}
