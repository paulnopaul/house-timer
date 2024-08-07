package tasks

import (
	"errors"
)

var ErrBadTaskMessage = errors.New("bad task message")

var ErrBadTaskEventType = errors.New("bad task event type")

var ErrBadTaskEventStep = errors.New("bad task event step")

var ErrBadTaskNumber = errors.New("bad task number")

var ErrBadTaskEvent = errors.New("bad task event")

var ErrNoTasks = errors.New("no tasks")

var ErrEventCollision = errors.New("event collision")

var ErrCreateTaskEvent = errors.New("failed to create task event")

var ErrCreateEmptyTask = errors.New("failed to create empty task")

var ErrAddTaskID = errors.New("failed to add task id")

var ErrCreateTaskName = errors.New("failed to create task name")

var ErrCreateTaskRegularity = errors.New("failed to create task regularity")

var ErrUpdateTaskStep = errors.New("failed to update task step")

var ErrParseRegularity = errors.New("failed to parse regularity")

var ErrFinishCreation = errors.New("failed to finish creation")

var ErrDeleteEvent = errors.New("failed to delete event")

var ErrUnknownTaskCreateStep = errors.New("unknown task create step")

var ErrUnknownTaskEditStep = errors.New("unknown task edit step")

var ErrGetCurrentTaskEvent = errors.New("failed to get current task event")

var ErrGetTasks = errors.New("failed to get tasks")

var ErrUpdateTask = errors.New("failed to update task")
