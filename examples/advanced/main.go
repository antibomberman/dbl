package advanced

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/antibomberman/DBL"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Подключение к базе данных
	// Connect to database
	db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/testdb?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	dbl := DBL.New("mysql", db)
	// Пример использования очередей
	// Queue usage example
	if err := queueExample(dbl); err != nil {
		log.Fatal(err)
	}
	// Пример использования событий
	// Events usage example
	if err := eventsExample(dbl); err != nil {
		log.Fatal(err)
	}
	// Пример геопространственных запросов
	// Geospatial queries example
	if err := geoExample(dbl); err != nil {
		log.Fatal(err)
	}
}

func queueExample(dbl *DBL.DBL) error {
	// Создаем отложенную операцию
	// Create a delayed operation
	err := dbl.Table("queued_operations").Queue(
		"send_email",
		map[string]interface{}{
			"to":      "user@example.com",
			"subject": "Hello!",
			"body":    "This is a delayed message",
		},
		time.Now().Add(1*time.Hour),
	)
	if err != nil {
		return err
	}
	// Обработка очереди
	// Process the queue
	return dbl.Table("queued_operations").ProcessQueue(func(op DBL.QueuedOperation) error {
		fmt.Printf("Processing operation: %s\n", op.Operation)
		return nil
	})
}

func eventsExample(dbl *DBL.DBL) error {
	// Регистрируем обработчики событий
	// Register event handlers
	qb := dbl.Table("users")

	qb.On(DBL.BeforeCreate, func(data interface{}) error {
		fmt.Println("Before creating user")
		return nil
	})
	qb.On(DBL.AfterCreate, func(data interface{}) error {
		fmt.Println("After creating user")
		return nil
	})
	// Создаем пользователя (сработают события)
	// Create a user (events will trigger)
	_, err := qb.Create(map[string]interface{}{
		"name":  "New User",
		"email": "new@example.com",
	})
	return err
}

func geoExample(dbl *DBL.DBL) error {
	// Поиск мест в радиусе 5 км от точки
	// Search for places within 5km radius from point
	var places []struct {
		ID   int64   `db:"id"`
		Name string  `db:"name"`
		Lat  float64 `db:"lat"`
		Lng  float64 `db:"lng"`
	}
	point := DBL.Point{
		Lat: 55.7558,
		Lng: 37.6173,
	}
	_, err := dbl.Table("places").
		GeoSearch("location", point, 5000). // радиус в метрах / radius in meters
		OrderBy("distance", "ASC").
		Limit(10).
		Get(&places)
	return err
}
