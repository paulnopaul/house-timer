package entities

import (
	"context"
	"fmt"
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
	ID           int64
	Name         string
	ChatID       int64
	Regularity   time.Duration
	LastReminded time.Time
	RemindAfter  time.Duration
}

func (u *UserTask) Recipient() string {
	return fmt.Sprintf("%d", u.ChatID)
}

type TaskStorage interface {
	CreateEmptyTask(ctx context.Context, chatID int64) (int64, error)
	CreateTaskName(ctx context.Context, taskID int64, taskName string) error
	CreateTaskRegularity(ctx context.Context, taskID int64, regularity time.Duration) error
	FinishCreation(ctx context.Context, taskID int64) error
	GetTasksForChat(ctx context.Context, chatID int64) ([]UserTask, error)
	UpdateTask(ctx context.Context, taskUpdate TaskUpdate) error
	GetChatIDs(ctx context.Context) ([]int64, error)
	DeleteTask(ctx context.Context, taskID int64) error
}

type TaskUpdate struct {
	TaskID       int64
	Name         *string
	RemindAfter  *time.Duration
	Regularity   *time.Duration
	LastReminded *time.Time
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

func NewGotEditNumberTaskResult() TaskMessageResult {
	return "GotEditNumberTaskResult"
}

func (t TaskMessageResult) IsGotEditNumberTaskResult() bool {
	return t == "GotEditNumberTaskResult"
}

func NewGotEditNameTaskResult() TaskMessageResult {
	return "GotEditNameTaskResult"
}

func (t TaskMessageResult) IsGotEditNameTaskResult() bool {
	return t == "GotEditNameTaskResult"
}

func NewGotEditRegularityTaskResult() TaskMessageResult {
	return "GotEditRegularityTaskResult"
}

func (t TaskMessageResult) IsGotEditRegularityTaskResult() bool {
	return t == "GotEditRegularityTaskResult"
}

func NewNeedRemindMessageResult() TaskMessageResult {
	return "NeedRemind"
}

func (t TaskMessageResult) IsNeedRemindMessageResult() bool {
	return t == "NeedRemind"
}

func NewNoRemindMessageResult() TaskMessageResult {
	return "NoRemind"
}

func (t TaskMessageResult) IsNoRemindMessageResult() bool {
	return t == "NoRemind"
}

type TaskUsecase interface {
	CreateEmptyTask(ctx context.Context, chatID int64) error
	HandleTaskMessage(ctx context.Context, chatID int64, message string) (TaskMessageResult, error)
	GetTasks(ctx context.Context, chatID int64) ([]UserTask, error)
	CurrentEventType(ctx context.Context, chatID int64) (TaskEventType, error)
	StartTaskEdit(ctx context.Context, chatID int64) error
	StartTaskNameEdit(ctx context.Context, chatID int64) error
	StartTaskRegularityEdit(ctx context.Context, chatID int64) error
	StopTaskEdit(ctx context.Context, chatID int64) error
	ResetTaskEdit(ctx context.Context, chatID int64) error
	RemindLater(ctx context.Context, chatID int64) error
	CompleteTask(ctx context.Context, chatID int64) error
	HandleRemind(ctx context.Context, chatID int64) (TaskMessageResult, error)
	DeleteCurrentTask(ctx context.Context, chatID int64) error
	StopTaskCreation(ctx context.Context, chatID int64) error
}
