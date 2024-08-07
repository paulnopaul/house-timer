package sqlite_repo

import (
	"context"
	"database/sql"
	"time"

	"house-timer/internal/pkg/entities"

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
	result, err := ts.db.Exec("UPDATE Tasks SET Name = ?, RemindedAt = ? WHERE ID = ?", taskName, time.Now().Unix(), taskID)
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
	rows, err := ts.db.Query("SELECT ID, Name, Regularity, RemindedAt, ChatID, RemindAfter  FROM Tasks WHERE ChatID = ? AND CreatedAt IS NOT NULL AND DeletedAt IS NULL ORDER BY CreatedAt ASC", chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []entities.UserTask
	for rows.Next() {
		var task entities.UserTask
		var regularitySeconds uint64
		var remindedSeconds int64
		var remindAfterSeconds int64
		if err := rows.Scan(&task.ID, &task.Name, &regularitySeconds, &remindedSeconds, &task.ChatID, &remindAfterSeconds); err != nil {
			return nil, err
		}
		task.Regularity = time.Duration(regularitySeconds) * time.Second
		task.LastReminded = time.Unix(remindedSeconds, 0)
		task.RemindAfter = time.Duration(remindAfterSeconds) * time.Second
		res = append(res, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (ts *SqliteTaskStorage) UpdateTask(ctx context.Context, update entities.TaskUpdate) error {
	tx, err := ts.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if update.Name != nil {
		_, err := tx.Exec("UPDATE Tasks SET Name = ? WHERE ID = ?", *update.Name, update.TaskID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	if update.RemindAfter != nil {
		_, err := tx.Exec("UPDATE Tasks SET RemindAfter = ? WHERE ID = ?", update.RemindAfter.Seconds(), update.TaskID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	if update.Regularity != nil {
		_, err := tx.Exec("UPDATE Tasks SET Regularity = ? WHERE ID = ?", update.Regularity.Seconds(), update.TaskID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	if update.LastReminded != nil {
		_, err := tx.Exec("UPDATE Tasks SET RemindedAt = ? WHERE ID = ?", update.LastReminded.Unix(), update.TaskID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (ts *SqliteTaskStorage) GetChatIDs(_ context.Context) ([]int64, error) {
	rows, err := ts.db.Query("SELECT DISTINCT ChatID from Tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []int64
	for rows.Next() {
		var chatID int64
		if err := rows.Scan(&chatID); err != nil {
			return nil, err
		}
		res = append(res, chatID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

// DeleteTask marks task as deleted
func (ts *SqliteTaskStorage) DeleteTask(_ context.Context, taskID int64) error {
	_, err := ts.db.Exec("UPDATE Tasks SET DeletedAt = ? WHERE ID = ?", time.Now().Unix(), taskID)
	if err != nil {
		return err
	}
	return nil
}
