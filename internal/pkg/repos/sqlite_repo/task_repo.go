package sqlite_repo

import (
	"context"
	"database/sql"
	"house-timer/internal/pkg/entities"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SqliteTaskStorage struct {
	db *sql.DB
}

func NewSqliteTaskStorage(db *sql.DB) *SqliteTaskStorage {
	return &SqliteTaskStorage{
		db: db,
	}
}

func (ts *SqliteTaskStorage) CreateEmptyTask(_ context.Context, chatID int64) (int64, error) {
	result, err := ts.db.Exec("INSERT INTO Tasks(ChatID) VALUES(?)", chatID)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (ts *SqliteTaskStorage) CreateTaskName(_ context.Context, taskID int64, taskName string) error {
	result, err := ts.db.Exec("UPDATE Tasks SET Name = ? WHERE ID = ?", taskName, taskID)
	if err != nil {
		return err
	}
	_, err = result.LastInsertId()
	if err != nil {
		return err
	}
	return nil
}

func (ts *SqliteTaskStorage) CreateTaskRegularity(_ context.Context, taskID int64, regularity time.Duration) error {
	result, err := ts.db.Exec("UPDATE Tasks SET Regularity = ? WHERE ID = ?", int64(regularity.Seconds()), taskID)
	if err != nil {
		return err
	}
	_, err = result.LastInsertId()
	if err != nil {
		return err
	}
	return nil
}

func (ts *SqliteTaskStorage) FinishCreation(_ context.Context, taskID int64) error {
	result, err := ts.db.Exec("UPDATE Tasks SET CreatedAt = ? WHERE ID = ?", time.Now().Unix(), taskID)
	if err != nil {
		return err
	}
	_, err = result.LastInsertId()
	if err != nil {
		return err
	}
	return nil
}

func (ts *SqliteTaskStorage) GetTasksForChat(_ context.Context, chatID int64) ([]entities.UserTask, error) {
	rows, err := ts.db.Query("SELECT ID, Name, Regularity FROM Tasks WHERE ChatID = ? AND CreatedAt IS NOT NULL AND DeletedAt IS NULL ORDER BY NAME ASC", chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []entities.UserTask
	for rows.Next() {
		var task entities.UserTask
		var regularitySeconds uint64
		if err := rows.Scan(&task.ID, &task.Name, &regularitySeconds); err != nil {
			return nil, err
		}
		task.Regularity = time.Duration(regularitySeconds) * time.Second
		res = append(res, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
