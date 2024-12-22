package basoc

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/antibomberman/dblayer"
)

// Структуры для работы с данными
type User struct {
	ID        int64     `db:"id"`
	Email     string    `db:"email"`
	Name      string    `db:"name"`
	Age       int       `db:"age"`
	CreatedAt time.Time `db:"created_at"`
	Version   int       `db:"version"`
}

type Post struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	Title     string    `db:"title"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}
type Comment struct {
	ID        int64     `db:"id"`
	PostID    int64     `db:"post_id"`
	UserID    int64     `db:"user_id"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}

func main() {
	// Подключение к базе данных
	db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/testdb?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// Инициализация DBLayer
	dbl := dblayer.NewDBLayer("mysql", db)
	// Создание таблиц
	if err := createTables(dbl); err != nil {
		log.Fatal(err)
	}
	// Примеры использования
	if err := examples(dbl); err != nil {
		log.Fatal(err)
	}
}

func createTables(dbl *dblayer.DBLayer) error {
	// Создание таблицы пользователей
	err := dbl.Create("users", func(schema *dblayer.Schema) {
		schema.ID()
		schema.String("email", 255).Unique()
		schema.String("name", 100)
		schema.Integer("age")
		schema.Timestamp("created_at")
		schema.Integer("version").Default(1)
	})
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы users: %w", err)
	}
	// Создание таблицы постов
	err = dbl.Create("posts", func(schema *dblayer.Schema) {
		schema.Integer("id").Primary().AutoIncrement()
		schema.Integer("user_id")
		schema.String("title", 200)
		schema.Text("content")
		schema.Timestamp("created_at")
		schema.ForeignKey("user_id", "users", "id")
	})
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы posts: %w", err)
	}
	// Создание таблицы комментариев
	err = dbl.Create("comments", func(schema *dblayer.Schema) {
		schema.Integer("id").Primary().AutoIncrement()
		schema.Integer("post_id")
		schema.Integer("user_id")
		schema.Text("content")
		schema.Timestamp("created_at")
		schema.ForeignKey("post_id", "posts", "id")
		schema.ForeignKey("user_id", "users", "id")
	})
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы comments: %w", err)
	}
	return nil
}

func examples(dbl *dblayer.DBLayer) error {
	// 1. Создание пользователя с транзакцией и аудитом
	tx, err := dbl.Begin()
	if err != nil {
		return err
	}
	user := &User{
		Email:     "agabek309@gmail.com",
		Name:      "Agabek Backend Developer",
		Age:       25,
		CreatedAt: time.Now(),
		Version:   1,
	}
	userID, err := tx.Table("users").
		WithAudit(1). // ID администратора
		Create(user)
	if err != nil {
		tx.Rollback()
		return err
	}
	// 2. Создание поста для пользователя
	post := &Post{
		UserID:    userID,
		Title:     "Мой первый пост",
		Content:   "Привет, мир!",
		CreatedAt: time.Now(),
	}
	postID, err := tx.Table("posts").Create(post)
	if err != nil {
		tx.Rollback()
		return err
	}
	// 3. Добавление комментария
	comment := &Comment{
		PostID:    postID,
		UserID:    userID,
		Content:   "Первый комментарий",
		CreatedAt: time.Now(),
	}
	_, err = tx.Table("comments").Create(comment)
	if err != nil {
		tx.Rollback()
		return err
	}
	// Фиксация транзакции
	if err := tx.Commit(); err != nil {
		return err
	}
	// 4. Сложный запрос с JOIN и кешированием
	var results []struct {
		UserName     string    `db:"name"`
		PostTitle    string    `db:"title"`
		CommentCount int       `db:"comment_count"`
		LastComment  time.Time `db:"last_comment"`
	}
	_, err = dbl.Table("users").
		Select("users.name, posts.title, COUNT(comments.id) as comment_count, MAX(comments.created_at) as last_comment").
		LeftJoin("posts", "posts.user_id = users.id").
		LeftJoin("comments", "comments.post_id = posts.id").
		GroupBy("users.id, posts.id").
		Having("comment_count > 0").
		OrderBy("last_comment", "DESC").
		Remember("user_posts_stats", 5*time.Minute).
		Get(&results)
	if err != nil {
		return err
	}
	// 5. Пакетное обновление с оптимистичной блокировкой
	users := []map[string]interface{}{
		{"id": 1, "age": 26, "version": 2},
		{"id": 2, "age": 31, "version": 2},
	}
	err = dbl.Table("users").BatchUpdate(users, "id", 100)
	if err != nil {
		return err
	}
	// 6. Полнотекстовый поиск
	var searchResults []Post
	_, err = dbl.Table("posts").
		Search([]string{"title", "content"}, "привет").
		OrderBy("search_rank", "DESC").
		Limit(10).
		Get(&searchResults)
	if err != nil {
		return err
	}
	// 7. Пагинация с метриками
	var pagedUsers []User
	collector := dblayer.NewMetricsCollector()
	result, err := dbl.Table("users").
		WithMetrics(collector).
		OrderBy("created_at", "DESC").
		Paginate(1, 10, &pagedUsers)
	if err != nil {
		return err
	}
	fmt.Printf("Всего пользователей: %d, Страниц: %d\n", result.Total, result.LastPage)
	return nil
}