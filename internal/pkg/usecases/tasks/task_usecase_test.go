package tasks

import (
	"context"
	"database/sql"
	"embed"
	"house-timer/internal/pkg/repos/sqlite_repo"
	"math/rand"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
)

//go:generate cp -r ../../../../migrations ./zz.generated_test_migrations

//go:embed zz.generated_test_migrations/*.sql
var embedMigrations embed.FS

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory SQLite database: %v", err)
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		panic(err)
	}

	if err := goose.Up(db, "zz.generated_test_migrations"); err != nil {
		panic(err)
	}
	return db
}

func generateChatID() int64 {
	return rand.Int63()
}

func TestTaskUsecase(t *testing.T) {
	db := setupTestDB(t)
	taskEventStorage := sqlite_repo.NewSqliteTaskEventStorage(db)
	taskStorage := sqlite_repo.NewSqliteTaskStorage(db)
	taskUsecase := NewTaskUsecase(taskStorage, taskEventStorage)

	chatID := generateChatID()
	ctx := context.Background()
	err := taskUsecase.CreateEmptyTask(ctx, chatID)
	assert.NoError(t, err)

	taskName := "NewTask1"
	res, err := taskUsecase.HandleTaskMessage(ctx, chatID, taskName)
	assert.NoError(t, err)
	assert.True(t, res.IsTaskNameCreated())

	res, err = taskUsecase.HandleTaskMessage(ctx, chatID, "2 дня")
	assert.NoError(t, err)
	assert.True(t, res.IsTaskCreated())

	tasks, err := taskStorage.GetTasksForChat(ctx, chatID)
	assert.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.Equal(t, tasks[0].Name, taskName)

	err = taskUsecase.CreateEmptyTask(ctx, chatID)
	assert.NoError(t, err)

	taskName = "NewTask2"
	res, err = taskUsecase.HandleTaskMessage(ctx, chatID, taskName)
	assert.NoError(t, err)
	assert.True(t, res.IsTaskNameCreated())

	res, err = taskUsecase.HandleTaskMessage(ctx, chatID, "2 дня")
	assert.NoError(t, err)
	assert.True(t, res.IsTaskCreated())

	tasks, err = taskStorage.GetTasksForChat(ctx, chatID)
	assert.NoError(t, err)
	assert.Len(t, tasks, 2)
	assert.Equal(t, tasks[1].Name, taskName)
}
