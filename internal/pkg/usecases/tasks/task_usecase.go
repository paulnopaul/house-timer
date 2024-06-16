package tasks

import (
	"context"
	"errors"
	"house-timer/internal/pkg/entities"
	"house-timer/pkg/regularity"
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
	eventID, err := t.tes.CreateTaskEvent(ctx, chatID)
	if err != nil {
		return errors.Join(NewErrCreateTaskEvent(chatID), err)
	}

	taskID, err := t.ts.CreateEmptyTask(ctx, chatID)
	if err != nil {
		return errors.Join(NewErrCreateEmptyTask(chatID), err)
	}

	err = t.tes.AddTaskID(ctx, eventID, taskID)
	if err != nil {
		return errors.Join(err, NewErrCreateTaskIDEvent(chatID, taskID))
	}
	return nil
}

func (t *TaskUsecase) HandleTaskMessage(ctx context.Context, chatID int64, message string) (entities.TaskMessageResult, error) {
	event, err := t.tes.GetCurrrentTaskEvent(ctx, chatID)
	if err != nil {
		return entities.NewEmptyTaskMessageResult(), errors.Join(NewErrGetCurrentTaskEvent(chatID), err)
	}

	if event.Type == entities.TaskCreationEvent {
		switch event.Step {
		case entities.TaskCreationWaitName:
			err := t.ts.CreateTaskName(ctx, event.TaskID, message)
			if err != nil {
				return entities.NewEmptyTaskMessageResult(), errors.Join(NewErrCreateTaskName(chatID, event.TaskID), err)
			}
			err = t.tes.MoveNext(ctx, event.ID)
			if err != nil {
				return entities.NewEmptyTaskMessageResult(), errors.Join(NewErrMoveNextTaskEvent(chatID, event.TaskID), err)
			}
			return entities.NewNameCreatedTaskResult(), nil
		case entities.TaskCreationWaitRegularity:
			regularity, err := regularity.ExtractRegularity(message)
			if err != nil {
				return entities.NewEmptyTaskMessageResult(), errors.Join(NewErrTaskRegularityParseError(chatID, event.TaskID), err)
			}
			err = t.ts.CreateTaskRegularity(ctx, event.TaskID, regularity)
			if err != nil {
				return entities.NewEmptyTaskMessageResult(), errors.Join(NewErrCreateTaskRegularityError(chatID, event.TaskID), err)
			}

			err = t.ts.FinishCreation(ctx, event.TaskID)
			if err != nil {
				return entities.NewEmptyTaskMessageResult(), errors.Join(NewErrCreateTaskRegularityError(chatID, event.TaskID), err)
			}

			err = t.tes.MoveNext(ctx, event.ID)
			if err != nil {
				return entities.NewEmptyTaskMessageResult(), errors.Join(NewErrMoveNextTaskEvent(chatID, event.TaskID), err)
			}
			return entities.NewCreatedTask(), nil
		default:
			return entities.NewEmptyTaskMessageResult(), NewErrUnknownTaskEventStep(chatID, event.TaskID, string(event.Step))
		}
	} else {
		return entities.NewEmptyTaskMessageResult(), NewErrUnknownTaskEventType(chatID, event.TaskID, string(event.Type))
	}
}

func (t *TaskUsecase) GetTasks(ctx context.Context, chatID int64) ([]entities.UserTask, error) {
	tasks, err := t.ts.GetTasksForChat(ctx, chatID)
	if err != nil {
		return nil, NewErrGetTasks(chatID)
	}
	return tasks, nil
}
