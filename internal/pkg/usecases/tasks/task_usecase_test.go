package tasks

import (
	"context"
	"database/sql"
	"embed"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
	"time"

	"house-timer/internal/pkg/repos/sqlite_repo"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

//go:generate rm -rf ./zz.generated_test_migrations
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
	require.NoError(t, err)

	taskName := "NewTask1"
	res, err := taskUsecase.HandleTaskMessage(ctx, chatID, taskName)
	require.NoError(t, err)
	require.True(t, res.IsTaskNameCreated())

	res, err = taskUsecase.HandleTaskMessage(ctx, chatID, "2 дня")
	require.NoError(t, err)
	require.True(t, res.IsTaskCreated())

	tasks, err := taskStorage.GetTasksForChat(ctx, chatID)
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	require.Equal(t, tasks[0].Name, taskName)
	require.Equal(t, tasks[0].Regularity, time.Hour*24*2)

	err = taskUsecase.CreateEmptyTask(ctx, chatID)
	require.NoError(t, err)

	taskName = "NewTask2"
	res, err = taskUsecase.HandleTaskMessage(ctx, chatID, taskName)
	require.NoError(t, err)
	require.True(t, res.IsTaskNameCreated())

	res, err = taskUsecase.HandleTaskMessage(ctx, chatID, "10 месяцев")
	require.NoError(t, err)
	require.True(t, res.IsTaskCreated())

	tasks, err = taskStorage.GetTasksForChat(ctx, chatID)
	require.NoError(t, err)
	require.Len(t, tasks, 2)
	require.Equal(t, tasks[1].Name, taskName)
	require.Equal(t, tasks[1].Regularity, time.Hour*24*30*10)

	err = taskUsecase.StartTaskEdit(ctx, chatID)
	require.NoError(t, err)
	res, err = taskUsecase.HandleTaskMessage(ctx, chatID, "1")
	require.NoError(t, err)
	require.True(t, res.IsGotEditNumberTaskResult())

	err = taskUsecase.StartTaskNameEdit(ctx, chatID)
	require.NoError(t, err)
	res, err = taskUsecase.HandleTaskMessage(ctx, chatID, "Новое имя")
	require.NoError(t, err)
	require.True(t, res.IsGotEditNameTaskResult())

	tasks, err = taskStorage.GetTasksForChat(ctx, chatID)
	require.NoError(t, err)
	require.Len(t, tasks, 2)
	require.Equal(t, tasks[0].Name, "Новое имя")

	err = taskUsecase.ResetTaskEdit(ctx, chatID)
	require.NoError(t, err)
	res, err = taskUsecase.HandleTaskMessage(ctx, chatID, "2")
	require.NoError(t, err)
	require.True(t, res.IsGotEditNumberTaskResult())

	err = taskUsecase.StartTaskNameEdit(ctx, chatID)
	require.NoError(t, err)
	res, err = taskUsecase.HandleTaskMessage(ctx, chatID, "Новое имя 2")
	require.NoError(t, err)
	require.True(t, res.IsGotEditNameTaskResult())

	tasks, err = taskStorage.GetTasksForChat(ctx, chatID)
	require.NoError(t, err)
	require.Len(t, tasks, 2)
	require.Equal(t, tasks[1].Name, "Новое имя 2")
}
