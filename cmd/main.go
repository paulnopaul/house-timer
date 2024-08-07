package main

import (
	"context"
	"database/sql"
	"embed"
	"house-timer/internal/pkg/remind"
	"log"
	"os"
	"time"

	"house-timer/internal/pkg/delivery"
	"house-timer/internal/pkg/repos/sqlite_repo"
	"house-timer/internal/pkg/usecases/tasks"

	"github.com/pressly/goose/v3"
	tele "gopkg.in/telebot.v3"
)

//go:generate rm -rf ./zz.generated_prod_migrations
//go:generate cp -r ../migrations ./zz.generated_prod_migrations
//go:embed zz.generated_prod_migrations/*.sql
var embedMigrations embed.FS

func setupTestDB() *sql.DB {
	// db, err := sql.Open("sqlite3", ":memory:")
	db, err := sql.Open("sqlite3", "tasks.sqlite")
	if err != nil {
		log.Fatalf("Failed to open in-memory SQLite database: %v", err)
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		panic(err)
	}

	if err := goose.Up(db, "zz.generated_prod_migrations"); err != nil {
		panic(err)
	}
	return db
}

func main() {
	pref := tele.Settings{
		Token:  os.Getenv("TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	db := setupTestDB()

	taskEventStorage := sqlite_repo.NewSqliteTaskEventStorage(db)
	taskStorage := sqlite_repo.NewSqliteTaskStorage(db)
	taskUsecase := tasks.NewTaskUsecase(taskStorage, taskEventStorage)
	delivery.NewDeliveryHandler(b, taskUsecase)

	r := remind.NewRemindHandler(taskStorage, taskUsecase, b)
	go r.Start(context.Background())
	b.Start()
}
