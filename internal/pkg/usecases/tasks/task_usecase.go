package tasks

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"house-timer/internal/pkg/entities"
	"house-timer/internal/pkg/repos/sqlite_repo"
	"house-timer/pkg/regularity"

	"github.com/go-logr/logr"
)

type TaskUsecase struct {
	ts  entities.TaskStorage
	tes entities.TaskEventStorage
}

func NewTaskUsecase(
	taskStorage entities.TaskStorage,
	taskEventStorage entities.TaskEventStorage,
) *TaskUsecase {
	return &TaskUsecase{
		ts:  taskStorage,
		tes: taskEventStorage,
	}
}

func (t *TaskUsecase) CreateEmptyTask(ctx context.Context, chatID int64) error {
	log, err := logr.FromContext(ctx)
	if err != nil {
		return err
	}

	// check if no current event
	_, err = t.tes.GetCurrentTaskEvent(ctx, chatID)
	if err == nil {
		return ErrEventCollision
	}

	log.Info("creating task event")
	eventID, err := t.tes.CreateTaskEvent(ctx, chatID, entities.TaskCreationEvent, entities.TaskCreationWaitName)
	if err != nil {
		return errors.Join(ErrCreateTaskEvent, err)
	}

	log.Info("creating empty task")
	taskID, err := t.ts.CreateEmptyTask(ctx, chatID)
	if err != nil {
		return errors.Join(ErrCreateEmptyTask, err)
	}

	log.Info("adding task id to event")
	err = t.tes.AddTaskID(ctx, eventID, taskID)
	if err != nil {
		return errors.Join(ErrAddTaskID, err)
	}
	return nil
}

func (t *TaskUsecase) handleCreateMessage(ctx context.Context, event *entities.UserTaskEvent, chatID int64, message string) (entities.TaskMessageResult, error) {
	switch event.Step {
	case entities.TaskCreationWaitName:
		err := t.ts.CreateTaskName(ctx, event.TaskID, message)
		if err != nil {
			return entities.NewEmptyTaskMessageResult(), errors.Join(ErrCreateTaskName, err)
		}
		err = t.tes.UpdateStep(ctx, chatID, entities.TaskCreationWaitRegularity)
		if err != nil {
			return entities.NewEmptyTaskMessageResult(), errors.Join(ErrUpdateTaskStep, err)
		}
		return entities.NewNameCreatedTaskResult(), nil
	case entities.TaskCreationWaitRegularity:
		reg, err := regularity.ExtractRegularity(message)
		if err != nil {
			return entities.NewEmptyTaskMessageResult(), errors.Join(ErrParseRegularity, err)
		}
		err = t.ts.CreateTaskRegularity(ctx, event.TaskID, reg)
		if err != nil {
			return entities.NewEmptyTaskMessageResult(), errors.Join(ErrCreateTaskRegularity, err)
		}

		err = t.ts.FinishCreation(ctx, event.TaskID)
		if err != nil {
			return entities.NewEmptyTaskMessageResult(), errors.Join(ErrFinishCreation, err)
		}

		err = t.tes.DeleteEvent(ctx, event.ID)
		if err != nil {
			return entities.NewEmptyTaskMessageResult(), errors.Join(ErrDeleteEvent, err)
		}
		return entities.NewCreatedTask(), nil
	default:
		return entities.NewEmptyTaskMessageResult(), ErrUnknownTaskCreateStep
	}
}

func (t *TaskUsecase) getOrderedTaskID(ctx context.Context, chatID int64, taskNum int64) (int64, error) {
	tasks, err := t.ts.GetTasksForChat(ctx, chatID)
	if err != nil {
		return 0, err
	}
	if taskNum <= 0 || int(taskNum) > len(tasks) {
		return 0, errors.Join(ErrBadTaskNumber, fmt.Errorf("wtf task number %d", taskNum))
	}
	return tasks[taskNum-1].ID, nil
}

func (t *TaskUsecase) handleUpdateMessage(ctx context.Context, event *entities.UserTaskEvent, chatID int64, message string) (entities.TaskMessageResult, error) {
	switch event.Step {
	case entities.TaskEditGetNumber:
		num, err := strconv.ParseInt(message, 10, 64)
		if err != nil {
			return entities.NewEmptyTaskMessageResult(), errors.Join(ErrBadTaskNumber, fmt.Errorf("cant parse '%s'", message))
		}
		taskID, err := t.getOrderedTaskID(ctx, chatID, num)
		if err != nil {
			return entities.NewEmptyTaskMessageResult(), err
		}
		err = t.tes.AddTaskID(ctx, event.ID, taskID)
		if err != nil {
			return entities.NewEmptyTaskMessageResult(), err
		}
		err = t.tes.UpdateStep(ctx, chatID, entities.TaskEditWait)
		if err != nil {
			return entities.NewEmptyTaskMessageResult(), err
		}
		return entities.NewGotEditNumberTaskResult(), nil
	case entities.TaskEditChangeName:
		event, err := t.tes.GetCurrentTaskEvent(ctx, chatID)
		if err != nil {
			return entities.NewEmptyTaskMessageResult(), err
		}
		err = t.ts.UpdateTask(ctx, entities.TaskUpdate{
			TaskID: event.TaskID,
			Name:   &message,
		})
		if err != nil {
			return entities.NewEmptyTaskMessageResult(), err
		}
		err = t.tes.UpdateStep(ctx, chatID, entities.TaskEditWait)
		if err != nil {
			return entities.NewEmptyTaskMessageResult(), err
		}
		return entities.NewGotEditNameTaskResult(), nil
	case entities.TaskEditChangeRegularity:
		event, err := t.tes.GetCurrentTaskEvent(ctx, chatID)
		if err != nil {
			return entities.NewEmptyTaskMessageResult(), errors.Join(ErrGetCurrentTaskEvent, err)
		}
		reg, err := regularity.ExtractRegularity(message)
		if err != nil {
			return entities.NewEmptyTaskMessageResult(), errors.Join(ErrParseRegularity, err)
		}
		err = t.ts.UpdateTask(ctx, entities.TaskUpdate{
			TaskID:     event.TaskID,
			Regularity: &reg,
		})
		if err != nil {
			return entities.NewEmptyTaskMessageResult(), errors.Join(ErrUpdateTask, err)
		}
		err = t.tes.UpdateStep(ctx, chatID, entities.TaskEditWait)
		if err != nil {
			return entities.NewEmptyTaskMessageResult(), errors.Join(ErrUpdateTaskStep, err)
		}
		return entities.NewGotEditRegularityTaskResult(), nil
	}
	return entities.NewEmptyTaskMessageResult(), ErrUnknownTaskEditStep
}

func (t *TaskUsecase) HandleTaskMessage(ctx context.Context, chatID int64, message string) (entities.TaskMessageResult, error) {
	event, err := t.tes.GetCurrentTaskEvent(ctx, chatID)
	if err != nil {
		return entities.NewEmptyTaskMessageResult(), errors.Join(ErrGetCurrentTaskEvent, err)
	}

	if event.Type == entities.TaskCreationEvent {
		return t.handleCreateMessage(ctx, &event, chatID, message)
	} else {
		return t.handleUpdateMessage(ctx, &event, chatID, message)
	}
}

func (t *TaskUsecase) GetTasks(ctx context.Context, chatID int64) ([]entities.UserTask, error) {
	tasks, err := t.ts.GetTasksForChat(ctx, chatID)
	if err != nil {
		return nil, errors.Join(ErrGetTasks, err)
	}
	return tasks, nil
}

func (t *TaskUsecase) CurrentEventType(ctx context.Context, chatID int64) (entities.TaskEventType, error) {
	event, err := t.tes.GetCurrentTaskEvent(ctx, chatID)
	if err != nil {
		return "", errors.Join(ErrGetCurrentTaskEvent, err)
	}
	return event.Type, nil
}

func (t *TaskUsecase) StartTaskEdit(ctx context.Context, chatID int64) error {
	_, err := t.tes.GetCurrentTaskEvent(ctx, chatID)
	if err == nil {
		return ErrEventCollision
	}

	tasks, err := t.ts.GetTasksForChat(ctx, chatID)
	if err != nil {
		return errors.Join(ErrCreateTaskEvent, err)
	}
	if len(tasks) == 0 {
		return ErrNoTasks
	}
	_, err = t.tes.CreateTaskEvent(ctx, chatID, entities.TaskEditEvent, entities.TaskEditGetNumber)
	if err != nil {
		return errors.Join(ErrCreateTaskEvent, err)
	}
	return nil
}

func (t *TaskUsecase) StartTaskNameEdit(ctx context.Context, chatID int64) error {
	currentEvent, err := t.tes.GetCurrentTaskEvent(ctx, chatID)
	if err != nil {
		return err
	}
	if currentEvent.Step != entities.TaskEditWait {
		return ErrBadTaskEvent
	}
	err = t.tes.UpdateStep(ctx, chatID, entities.TaskEditChangeName)
	if err != nil {
		return errors.Join(ErrCreateTaskEvent, err)
	}
	return nil
}

func (t *TaskUsecase) StartTaskRegularityEdit(ctx context.Context, chatID int64) error {
	currentEvent, err := t.tes.GetCurrentTaskEvent(ctx, chatID)
	if err != nil {
		return err
	}
	if currentEvent.Step != entities.TaskEditWait {
		return ErrBadTaskEvent
	}
	err = t.tes.UpdateStep(ctx, chatID, entities.TaskEditChangeRegularity)
	if err != nil {
		return errors.Join(ErrUpdateTaskStep, err)
	}
	return nil
}

func (t *TaskUsecase) StopTaskEdit(ctx context.Context, chatID int64) error {
	currentEvent, err := t.tes.GetCurrentTaskEvent(ctx, chatID)
	if err != nil {
		return err
	}
	if (currentEvent.Step != entities.TaskEditGetNumber) && (currentEvent.Step != entities.TaskEditWait) {
		return ErrBadTaskEvent
	}
	err = t.tes.DeleteEvent(ctx, currentEvent.ID)
	if err != nil {
		return errors.Join(ErrDeleteEvent, err)
	}
	return nil
}

func (t *TaskUsecase) ResetTaskEdit(ctx context.Context, chatID int64) error {
	currentEvent, err := t.tes.GetCurrentTaskEvent(ctx, chatID)
	if err != nil {
		return err
	}
	if currentEvent.Step != entities.TaskEditWait {
		return ErrBadTaskEvent
	}
	err = t.tes.UpdateStep(ctx, chatID, entities.TaskEditGetNumber)
	if err != nil {
		return errors.Join(ErrUpdateTaskStep, err)
	}
	return nil
}

func (t *TaskUsecase) RemindLater(ctx context.Context, chatID int64) error {
	taskEvent, err := t.tes.GetCurrentTaskEvent(ctx, chatID)
	if err != nil {
		return err
	}
	if taskEvent.Step != entities.TaskRemindWait || taskEvent.Type != entities.TaskRemindEvent {
		return ErrBadTaskEvent
	}
	remindDur := time.Hour * 24
	err = t.ts.UpdateTask(ctx, entities.TaskUpdate{
		TaskID:      taskEvent.TaskID,
		RemindAfter: &remindDur,
	})
	if err != nil {
		return errors.Join(ErrUpdateTask, err)
	}
	err = t.tes.DeleteEvent(ctx, taskEvent.ID)
	if err != nil {
		return err
	}
	return nil
}

func (t *TaskUsecase) CompleteTask(ctx context.Context, chatID int64) error {
	taskEvent, err := t.tes.GetCurrentTaskEvent(ctx, chatID)
	if err != nil {
		return err
	}
	if taskEvent.Step != entities.TaskRemindWait || taskEvent.Type != entities.TaskRemindEvent {
		return ErrBadTaskEvent
	}
	now := time.Now()
	remindAfter := time.Duration(0)
	err = t.ts.UpdateTask(ctx, entities.TaskUpdate{
		TaskID:       taskEvent.TaskID,
		LastReminded: &now,
		RemindAfter:  &remindAfter,
	})
	if err != nil {
		return errors.Join(ErrUpdateTask, err)
	}
	err = t.tes.DeleteEvent(ctx, taskEvent.ID)
	if err != nil {
		return err
	}
	return nil
}

// DeleteCurrentTask
func (t *TaskUsecase) DeleteCurrentTask(ctx context.Context, chatID int64) error {
	currentEvent, err := t.tes.GetCurrentTaskEvent(ctx, chatID)
	if err != nil {
		return err
	}
	if currentEvent.Step != entities.TaskEditWait {
		return ErrBadTaskEvent
	}
	err = t.ts.DeleteTask(ctx, currentEvent.TaskID)
	if err != nil {
		return err
	}
	err = t.tes.UpdateStep(ctx, chatID, entities.TaskEditWait)
	if err != nil {
		return err
	}
	return nil
}

func (t *TaskUsecase) HandleRemind(ctx context.Context, chatID int64) (entities.TaskMessageResult, error) {
	_, err := t.tes.GetCurrentTaskEvent(ctx, chatID)
	if err != nil {
		if errors.Is(err, sqlite_repo.ErrNoTaskEvent) {
			_, err = t.tes.CreateTaskEvent(ctx, chatID, entities.TaskRemindEvent, entities.TaskRemindWait)
			if err != nil {
				return entities.NewEmptyTaskMessageResult(), errors.Join(ErrCreateTaskEvent, err)
			}
			return entities.NewNeedRemindMessageResult(), nil
		} else {
			return entities.NewEmptyTaskMessageResult(), errors.Join(ErrGetCurrentTaskEvent, err)
		}
	}
	return entities.NewNoRemindMessageResult(), nil
}

func (t *TaskUsecase) StopTaskCreation(ctx context.Context, chatID int64) error {
	currentEvent, err := t.tes.GetCurrentTaskEvent(ctx, chatID)
	if err != nil {
		return err
	}
	if currentEvent.Step != entities.TaskCreationWaitName && currentEvent.Step != entities.TaskCreationWaitRegularity {
		return ErrBadTaskEvent
	}
	err = t.tes.DeleteEvent(ctx, currentEvent.ID)
	if err != nil {
		return err
	}
	err = t.ts.DeleteTask(ctx, currentEvent.TaskID)
	if err != nil {
		return err
	}
	return nil
}
