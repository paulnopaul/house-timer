package entities

import (
	"context"
)

type TaskEventType string

type TaskEventStep string

const (
	TaskCreationEvent TaskEventType = "task_create"
	TaskEditEvent     TaskEventType = "task_edit"

	TaskRemindEvent TaskEventType = "task_remind"
)

func (t TaskEventStep) GetType() TaskEventType {
	switch t {
	case TaskCreationWaitName:
	case TaskCreationWaitRegularity:
	case TaskCreationCompleted:
		return TaskCreationEvent
	}
	return TaskEditEvent
}

const (
	TaskCreationWaitName       TaskEventStep = "task_creation_wait_name"
	TaskCreationWaitRegularity TaskEventStep = "task_creation_wait_regularity"
	TaskCreationCompleted      TaskEventStep = "task_creation_completed"

	TaskEditGetNumber        TaskEventStep = "task_edit_get_number"
	TaskEditChangeName       TaskEventStep = "task_edit_wait_name"
	TaskEditWait             TaskEventStep = "task_edit_wait"
	TaskEditChangeRegularity TaskEventStep = "task_edit_wait_regularity"
	TaskEditCompleted        TaskEventStep = "task_edit_completed"

	TaskRemindWait TaskEventStep = "task_remind_wait"
)

// TaskEvent only one active task_event per chat
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
	CreateTaskEvent(ctx context.Context, chatID int64, event TaskEventType, step TaskEventStep) (int64, error)
	AddTaskID(ctx context.Context, eventID int64, taskID int64) error
	GetCurrentTaskEvent(ctx context.Context, chatID int64) (UserTaskEvent, error)
	DeleteEvent(ctx context.Context, eventID int64) error
	UpdateStep(ctx context.Context, chatID int64, newStep TaskEventStep) error
}
