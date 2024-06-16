package entities

import (
	"context"
	"time"
)

type DBEntity struct {
	ID        int64
	CreatedAt time.Time
	DeletedAt time.Time
}

type Task struct {
	DBEntity

	ChatID int64

	Name        string
	Regulatiry  time.Duration
	LastUpdated time.Time
}

type UserTask struct {
	ID         int64
	Name       string
	Regularity time.Duration
}

// type Check struct {
// 	DBEntity

// 	TaskID       int64
// 	Name         int64
// 	RemindBefore time.Duration
// }

type TaskStorage interface {
	CreateEmptyTask(ctx context.Context, chatID int64) (int64, error)
	CreateTaskName(ctx context.Context, taskID int64, taskName string) error
	CreateTaskRegularity(ctx context.Context, taskID int64, regularity time.Duration) error
	FinishCreation(ctx context.Context, taskID int64) error
	GetTasksForChat(ctx context.Context, chatID int64) ([]UserTask, error)
}

type TaskMessageResult string

func NewEmptyTaskMessageResult() TaskMessageResult {
	return ""
}

func NewNameCreatedTaskResult() TaskMessageResult {
	return "TaskNameCreated"
}

func (t TaskMessageResult) IsTaskNameCreated() bool {
	return t == "TaskNameCreated"
}

func NewCreatedTask() TaskMessageResult {
	return "TaskCreated"
}

func (t TaskMessageResult) IsTaskCreated() bool {
	return t == "TaskCreated"
}

type TaskUsecase interface {
	CreateEmptyTask(ctx context.Context, chatID int64) error
	HandleTaskMessage(ctx context.Context, chatID int64, message string) (TaskMessageResult, error)
	GetTasks(ctx context.Context, chatID int64) ([]UserTask, error)
}

type TaskEventType string

const (
	TaskCreationEvent TaskEventType = "TaskCreation"
)

type TaskEventStep string

const (
	TaskCreationWaitName       TaskEventStep = "task_creation_wait_name"
	TaskCreationWaitRegularity TaskEventStep = "task_creation_wait_regularity"
	TaskCreationCompleted      TaskEventStep = "task_creation_completed"
)

// only one active task_event per chat
type TaskEvent struct {
	DBEntity

	Type   TaskEventType
	Step   TaskEventStep
	TaskID int64
	ChatID int64
}

type UserTaskEvent struct {
	ID     int64
	Type   TaskEventType
	Step   TaskEventStep
	TaskID int64
	ChatID int64
}

type TaskEventStorage interface {
	CreateTaskEvent(ctx context.Context, chatID int64) (int64, error)
	AddTaskID(ctx context.Context, eventID int64, taskID int64) error
	GetCurrrentTaskEvent(ctx context.Context, chatID int64) (UserTaskEvent, error)
	MoveNext(ctx context.Context, chatID int64) error
}
