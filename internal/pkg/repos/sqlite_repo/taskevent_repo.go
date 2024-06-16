package sqlite_repo

import (
	"context"
	"database/sql"
	"errors"
	"house-timer/internal/pkg/entities"
	"time"
)

type SqliteTaskEventStorage struct {
	db *sql.DB
}

func NewSqliteTaskEventStorage(db *sql.DB) *SqliteTaskEventStorage {
	return &SqliteTaskEventStorage{
		db: db,
	}
}

func (ts *SqliteTaskEventStorage) CreateTaskEvent(_ context.Context, chatID int64) (int64, error) {
	result, err := ts.db.Exec("INSERT INTO TaskEvents(ChatID, CreatedAt, Type, Step) VALUES(?, ?, ?, ?)",
		chatID,
		time.Now().Unix(),
		entities.TaskCreationEvent,
		entities.TaskCreationWaitName)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (ts *SqliteTaskEventStorage) AddTaskID(_ context.Context, eventID int64, taskID int64) error {
	result, err := ts.db.Exec("UPDATE TaskEvents SET TaskID = ? WHERE ID = ?", taskID, eventID)
	if err != nil {
		return err
	}
	_, err = result.LastInsertId()
	if err != nil {
		return err
	}
	return nil
}

var ErrNoTaskEvent = errors.New("ErrNoTasks")

func (ts *SqliteTaskEventStorage) GetCurrrentTaskEvent(_ context.Context, chatID int64) (entities.UserTaskEvent, error) {
	rows, err := ts.db.Query(
		`SELECT ID, Type, Step, TaskID, ChatID 
		FROM TaskEvents 
		WHERE ChatID = ? AND DeletedAt IS NULL AND CreatedAt IS NOT NULL`,
		chatID)
	if err != nil {
		return entities.UserTaskEvent{}, err
	}
	defer rows.Close()
	var res []entities.UserTaskEvent
	for rows.Next() {
		var task entities.UserTaskEvent
		var taskID sql.NullInt64
		if err := rows.Scan(&task.ID, &task.Type, &task.Step, &taskID, &task.ChatID); err != nil {
			return entities.UserTaskEvent{}, err
		}
		if taskID.Valid {
			task.TaskID = taskID.Int64
		}
		res = append(res, task)
	}
	if err := rows.Err(); err != nil {
		return entities.UserTaskEvent{}, err
	}
	if len(res) == 0 {
		return entities.UserTaskEvent{}, ErrNoTaskEvent
	}
	if len(res) > 1 {
		return entities.UserTaskEvent{}, errors.New("Ты че ахуел, я же сказал ТОЛЬКО ОДНО СУКА СОБЫТИЕ ТАСКА НА ЧАТ")
	}
	return res[0], nil
}

var TaskEventStepFlow = map[entities.TaskEventStep]entities.TaskEventStep{
	entities.TaskCreationWaitName:       entities.TaskCreationWaitRegularity,
	entities.TaskCreationWaitRegularity: entities.TaskCreationCompleted,
}

func (ts *SqliteTaskEventStorage) MoveNext(_ context.Context, eventID int64) error {
	tx, err := ts.db.Begin()
	if err != nil {
		return err
	}
	row := tx.QueryRow(
		`SELECT Step
		FROM TaskEvents
		WHERE ID = ?`,
		eventID)
	if err := row.Err(); err != nil {
		tx.Rollback()
		return err
	}
	var step entities.TaskEventStep
	if err := row.Scan(&step); err != nil {
		tx.Rollback()
		return err
	}

	nextStep, ok := TaskEventStepFlow[step]
	if !ok {
		tx.Rollback()
		return errors.New("чета непонятный шаг евента какой-то")
	}

	_, err = tx.Exec(
		`UPDATE TaskEvents
		SET Step = ?
		WHERE ID = ?`,
		nextStep, eventID)

	if err != nil {
		tx.Rollback()
		return err
	}
	_, isNotLast := TaskEventStepFlow[nextStep]
	if !isNotLast {
		_, err = tx.Exec(
			`UPDATE TaskEvents
			SET DeletedAt = ?
			WHERE ID = ?`,
			time.Now().Unix(), eventID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
