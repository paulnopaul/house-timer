package tasks

import "fmt"

// TaskEventStorageError
// --------------------------------------

type ErrEvent struct {
	chatID int64
}

func (e *ErrEvent) Error() string {
	return fmt.Sprintf("event for chat %d", e.chatID)
}

type ErrCreateTaskEvent struct {
	ErrEvent
}

func (e *ErrCreateTaskEvent) Error() string {
	return fmt.Sprintf("failed creating event: %s", e.ErrEvent.Error())
}

func NewErrCreateTaskEvent(chatID int64) error {
	return &ErrCreateTaskEvent{ErrEvent{chatID: chatID}}
}

type ErrGetCurrentTaskEvent struct {
	ErrEvent
}

func (e *ErrGetCurrentTaskEvent) Error() string {
	return fmt.Sprintf("failed to get event: %s", e.ErrEvent.Error())
}

func NewErrGetCurrentTaskEvent(chatID int64) error {
	return &ErrGetCurrentTaskEvent{ErrEvent{chatID: chatID}}
}

type ErrCreateTaskIDEvent struct {
	ErrEvent
	taskID int64
}

func (e *ErrCreateTaskIDEvent) Error() string {
	return fmt.Sprintf("failed to add task id %d: %s", e.taskID, e.ErrEvent.Error())
}

func NewErrCreateTaskIDEvent(chatID int64, taskID int64) error {
	err := &ErrCreateTaskIDEvent{
		taskID: taskID,
	}
	err.chatID = chatID
	return err
}

type ErrMoveNextTaskEvent struct {
	ErrEvent
	taskID int64
}

func (e *ErrMoveNextTaskEvent) Error() string {
	return fmt.Sprintf("failed to move next for task %d: %s", e.taskID, e.ErrEvent.Error())
}

func NewErrMoveNextTaskEvent(chatID int64, taskID int64) error {
	err := &ErrMoveNextTaskEvent{
		taskID: taskID,
	}
	err.chatID = chatID
	return err
}

type ErrUnknownTaskEventStep struct {
	ErrEvent
	taskID int64
	step   string
}

func (e *ErrUnknownTaskEventStep) Error() string {
	return fmt.Sprintf("unknown next step '%s' for task %d: %s", e.step, e.taskID, e.ErrEvent.Error())
}

func NewErrUnknownTaskEventStep(chatID int64, taskID int64, step string) error {
	err := &ErrUnknownTaskEventStep{
		taskID: taskID,
	}
	err.chatID = chatID
	err.step = step
	return err
}

type ErrUnknownTaskEventType struct {
	ErrEvent
	taskID    int64
	eventType string
}

func (e *ErrUnknownTaskEventType) Error() string {
	return fmt.Sprintf("unknown next step '%s' for task %d: %s", e.eventType, e.taskID, e.ErrEvent.Error())
}

func NewErrUnknownTaskEventType(chatID int64, taskID int64, eventType string) error {
	err := &ErrUnknownTaskEventType{
		taskID: taskID,
	}
	err.chatID = chatID
	err.eventType = eventType
	return err
}

// TaskStorageError
// --------------------------------------

type ErrTask struct {
	chatID int64
	taskID int64
}

func (e *ErrTask) Error() string {
	if e.taskID != 0 {
		return fmt.Sprintf("task %d for chat %d", e.taskID, e.chatID)
	}
	return fmt.Sprintf("task error for chat %d", e.chatID)
}

type ErrCreateEmptyTask struct {
	ErrTask
}

func (e *ErrCreateEmptyTask) Error() string {
	return fmt.Sprintf("failed creating task: %s", e.ErrTask.Error())
}

func NewErrCreateEmptyTask(chatID int64) error {
	return &ErrCreateEmptyTask{ErrTask{chatID: chatID}}
}

type ErrCreateTaskName struct {
	ErrTask
}

func (e *ErrCreateTaskName) Error() string {
	return fmt.Sprintf("failed creating task name: %s", e.ErrTask.Error())
}

func NewErrCreateTaskName(chatID int64, taskID int64) error {
	return &ErrCreateTaskName{ErrTask{chatID: chatID, taskID: taskID}}
}

type ErrParseTaskRegularity struct {
	ErrTask
}

func (e *ErrParseTaskRegularity) Error() string {
	return fmt.Sprintf("failed parsing task regularity: %s", e.ErrTask.Error())
}

func NewErrTaskRegularityParseError(chatID int64, taskID int64) error {
	return &ErrParseTaskRegularity{ErrTask{chatID: chatID, taskID: taskID}}
}

type ErrCreateTaskRegularity struct {
	ErrTask
}

func (e *ErrCreateTaskRegularity) Error() string {
	return fmt.Sprintf("failed creating task regularity: %s", e.ErrTask.Error())
}

func NewErrCreateTaskRegularityError(chatID int64, taskID int64) error {
	return &ErrCreateTaskRegularity{ErrTask{chatID: chatID, taskID: taskID}}
}

type ErrFinishTaskCreation struct {
	ErrTask
}

func (e *ErrFinishTaskCreation) Error() string {
	return fmt.Sprintf("failed finishing task creation: %s", e.ErrTask.Error())
}

func NewErrFinishTaskCreation(chatID int64, taskID int64) error {
	return &ErrFinishTaskCreation{ErrTask{chatID: chatID, taskID: taskID}}
}


type ErrGetTasks struct {
	ErrTask
}

func (e *ErrGetTasks) Error() string {
	return fmt.Sprintf("failed to get tasks: %s", e.ErrTask.Error())
}

func NewErrGetTasks(chatID int64) error {
	return &ErrGetTasks{ErrTask{chatID: chatID}}
}
