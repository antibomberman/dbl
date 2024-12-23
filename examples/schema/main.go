package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/antibomberman/dblayer"
)

var DBLayer *dblayer.DBLayer

func main() {

	db, err := sql.Open("sqlite", "./examples/example.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	DBLayer = dblayer.New("sqlite", db)

}

// CreateTable создает таблицу
func CreateTable() {
	DBLayer.CreateTable("users", func(table *dblayer.Schema) {
		table.BigIncrements("id")
		table.String("name", 255)
		table.String("email", 255)
		table.Timestamps()
	})
}
func UpdateTable() {
	// Обновление таблицы
	DBLayer.Update("users", func(table *dblayer.Schema) {
		// Добавление новых колонок - тот же API, что и при создании
		table.String("phone", 20)
		// Специфичные для обновления операции
		table.RenameColumn("name", "full_name")
		table.DropColumn("old_field")
		table.ModifyColumn(dblayer.Column{
			Name:   "email",
			Type:   "VARCHAR",
			Length: 255,
			Unique: true,
		})
	})
}

func BuildTable() {
	// Создание таблицы пользователей
	userTable := DBLayer.CreateTable("users", func(table *dblayer.Schema) {
		table.Column("id").Type("bigint").AutoIncrement().Primary().Add()
		table.Column("name").Type("varchar", 255).Comment("Имя пользователя").Add()
		table.Column("email").Type("varchar", 255).Unique().Add()
		table.Column("password").Type("varchar", 255).Add()
		table.Column("status").Type("enum", 20).Default("active").Add()
		table.Column("created_at").Type("timestamp").Default("CURRENT_TIMESTAMP").Add()
		table.Column("updated_at").Type("timestamp").Nullable().Add()
		
		table.UniqueKey("uk_email", "email")
		table.Index("idx_status", "status")
		table.Comment("Таблица пользователей")

	sql := userTable.Build()
	fmt.Println(sql)

	// Создание таблицы заказов с внешними ключами
	orderTable := DBLayer.CreateTable("orders", func(table *dblayer.Schema) {
		table.Column("id").Type("bigint").AutoIncrement().Primary().Add()
		table.Column("user_id").Type("bigint").Add()
		table.Column("total").Type("decimal", 10).Default(0).Add()
		table.Column("status").Type("varchar", 50).Default("pending").Add()
		table.Column("created_at").Type("timestamp").Default("CURRENT_TIMESTAMP").Add()
		table.ForeignKey("user_id", "users", "id").
			OnDelete("CASCADE").
			OnUpdate("CASCADE").
			Add()
		table.Index("idx_user", "user_id")
		table.Index("idx_status", "status")

	sql = orderTable.Build()
	fmt.Println(sql)

	// Создание таблицы с составным ключом
	membershipTable := DBLayer.Builder("team_members").
		Column("team_id").Type("bigint").Add().
		Column("user_id").Type("bigint").Add().
		Column("role").Type("varchar", 50).Add().
		PrimaryKey("team_id", "user_id").
		ForeignKey("team_id", "teams", "id").
		OnDelete("CASCADE").
		Add().
		ForeignKey("user_id", "users", "id").
		OnDelete("CASCADE").
		Add()

	sql = membershipTable.Build()
	fmt.Println(sql)
}

func AlterTable() {
	DBLayer.Transaction(func(tx *dblayer.Transaction) error {
		// Добавление колонок
		alter := DBLayer.Alter("users").
			AddColumn(dblayer.Column{
				Name:     "phone",
				Type:     "varchar",
				Length:   20,
				Nullable: true,
				After:    "email",
			}).
			AddColumn(dblayer.Column{
				Name:    "age",
				Type:    "int",
				Default: 0,
			})

		// Изменение колонок
		alter.ModifyColumn(dblayer.Column{
			Name:    "status",
			Type:    "enum",
			Length:  30,
			Default: "active",
		})

		// Переименование колонки
		alter.RenameColumn("name", "full_name")

		// Добавление индексов
		alter.AddIndex("idx_phone", []string{"phone"}, false)
		alter.AddIndex("uk_phone", []string{"phone", "email"}, true)

		// Добавление внешнего ключа
		alter.AddForeignKey("fk_role", "role_id", dblayer.ForeignKey{
			Table:    "roles",
			Column:   "id",
			OnDelete: "SET NULL",
			OnUpdate: "CASCADE",
		})

		// Изменение параметров таблицы
		alter.ChangeEngine("MyISAM").
			ChangeCharset("utf8mb4", "utf8mb4_unicode_ci")

		// Выполнение изменений
		return alter.Execute()
	})
}

func DropTable() {
	// Удаление таблицы
	DBLayer.Drop("users", "orders").
		IfExists().
		Cascade().
		Execute()

	// Удаление временной таблицы
	DBLayer.Drop("temp_users").
		Temporary().
		Execute()

	// Неблокирующее удаление (PostgreSQL)
	DBLayer.Drop("large_table").
		Concurrent().
		Execute()

	// Принудительное удаление (MySQL)
	DBLayer.Drop("locked_table").
		Force().
		Execute()

}
func DropTruncate() {
	DBLayer.Transaction(func(tx *dblayer.Transaction) error {

		// Очистка таблицы со сбросом автоинкремента
		DBLayer.Truncate("orders").
			RestartIdentity().
			Cascade().
			Execute()

		// Очистка нескольких таблиц
		DBLayer.Truncate("orders", "order_items").
			Cascade().
			Execute()

		// Очистка с продолжением автоинкремента (PostgreSQL)
		DBLayer.Truncate("products").
			ContinueIdentity().
			Execute()

		// Принудительная очистка (MySQL)
		DBLayer.Truncate("locked_table").
			Force().
			Execute()

		return nil
	})
}
