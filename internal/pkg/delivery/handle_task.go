package delivery

import (
	"context"
	"errors"
	"fmt"
	"house-timer/internal/pkg/logmw"
	"house-timer/internal/pkg/repos/sqlite_repo"
	"house-timer/internal/pkg/usecases/tasks"
	"log"
	"log/slog"
	"time"

	"house-timer/internal/pkg/entities"

	"github.com/go-logr/logr"
	tele "gopkg.in/telebot.v3"
)

type deliveryHandler struct {
	mainMenu           *tele.ReplyMarkup
	taskEditMenu       *tele.ReplyMarkup
	taskEditMenuGoBack *tele.ReplyMarkup
	taskCreateStopMenu *tele.ReplyMarkup
	logger             logr.Logger

	taskUsecase entities.TaskUsecase
}

func NewDeliveryHandler(bot *tele.Bot, taskUsecase entities.TaskUsecase) {
	mainMenu := &tele.ReplyMarkup{}
	btnNewTask := mainMenu.Data("Создад", "createTask")
	btnEditTask := mainMenu.Data("Изменит", "editTask")
	mainMenu.Inline(
		mainMenu.Row(btnNewTask),
		mainMenu.Row(btnEditTask),
	)

	taskEditMenu := &tele.ReplyMarkup{}
	btnEditName := taskEditMenu.Data("Изменить название", "editTaskName")
	btnEditRegularity := taskEditMenu.Data("Изменить регулярность", "editTaskRegularity")
	btnDeleteTask := taskEditMenu.Data("Удалить задачу", "editDeleteTask")
	btnEditGoBack := taskEditMenu.Data("Изменить другую задачу", "taskEditAnother")
	btnEditStop := taskEditMenu.Data("Закончить изменение задач", "taskEditStop")
	taskEditMenu.Inline(
		taskEditMenu.Row(btnEditName),
		taskEditMenu.Row(btnEditRegularity),
		taskEditMenu.Row(btnDeleteTask),
		taskEditMenu.Row(btnEditGoBack),
		taskEditMenu.Row(btnEditStop),
	)

	taskEditMenuGoBack := &tele.ReplyMarkup{}
	taskEditMenuGoBack.Inline(
		taskEditMenuGoBack.Row(btnEditStop),
	)

	btnCreateStop := taskEditMenu.Data("Галя, отмена!", "taskCreateStop")

	taskCreateStopMenu := &tele.ReplyMarkup{}
	taskCreateStopMenu.Inline(
		taskCreateStopMenu.Row(btnCreateStop),
	)

	dh := deliveryHandler{
		mainMenu:           mainMenu,
		taskEditMenu:       taskEditMenu,
		taskEditMenuGoBack: taskEditMenuGoBack,
		taskCreateStopMenu: taskCreateStopMenu,
		logger:             logr.FromSlogHandler(slog.NewTextHandler(log.Writer(), nil)),

		taskUsecase: taskUsecase,
	}

	bot.Use(logmw.NewLogMW(dh.logger))
	bot.Handle("/start", dh.handleStart)
	bot.Handle(&btnNewTask, dh.handleNewTask)
	bot.Handle(&btnEditTask, dh.handleEditTask)

	bot.Handle(&btnEditName, dh.handleEditTaskName)
	bot.Handle(&btnEditRegularity, dh.handleEditTaskRegularity)
	bot.Handle(&btnDeleteTask, dh.handleDeleteTask)

	bot.Handle(&btnEditGoBack, dh.handleEditGoBack)
	bot.Handle(&btnEditStop, dh.handleEditStop)

	bot.Handle(&btnCreateStop, dh.handleCreateStop)

	bot.Handle(tele.OnText, dh.handleMessages)
}

const internalError = "Что-то пошло не так, обратитесь к @paulnopaul"
const unknownAction = "Я не понимаю, чего вы хотите, начните c создания задачи или изменения существующей"

func (dh deliveryHandler) handleStart(c tele.Context) error {
	chatID := c.Chat().ID
	return c.Send(fmt.Sprint("Привет, я бот-напоминалка редких, но очень нужных задач :)", chatID), dh.mainMenu)
}

func (dh deliveryHandler) handleNewTask(c tele.Context) error {
	chatID := c.Chat().ID

	log := logmw.GetLogger(c)
	log.Info("Creating task")
	ctx := logr.NewContext(context.Background(), log)
	log.Info("creating empty task")
	err := dh.taskUsecase.CreateEmptyTask(ctx, chatID)
	if err != nil {
		if errors.Is(err, tasks.ErrEventCollision) {
			log.Info("event collision")
			return c.Send("Надо закончить предыдущее действие, чтобы создать задачу")
		}
		log.Error(err, "failed to create empty task")
		return c.Send(internalError)
	}
	log.Info("empty task created")
	return c.Send("О чем надо напоминать?", dh.taskCreateStopMenu)
}

func (dh deliveryHandler) handleEditTask(c tele.Context) error {
	chatID := c.Chat().ID
	log := logmw.GetLogger(c)
	ctx := logr.NewContext(context.Background(), log)

	err := dh.taskUsecase.StartTaskEdit(ctx, chatID)
	if err != nil {
		if errors.Is(err, tasks.ErrEventCollision) {
			return c.Send("Надо закончить предыдущее действие, чтобы редактировать задачу")
		} else if errors.Is(err, tasks.ErrNoTasks) {
			return c.Send("Надо сначала создать задачи, чтобы их менять, ы", dh.mainMenu)
		}
		log.Error(err, "failed to start task edit")
		return c.Send(internalError)
	}

	chatTasks, err := dh.taskUsecase.GetTasks(ctx, chatID)
	if err != nil {
		log.Error(err, "failed to get tasks")
		return c.Send("Что-то пошло не так, почитай там логи что ли, лох")
	}
	return c.Send(formatTasks(chatTasks)+"Какую задачку будем менять? (введи номер)", dh.taskEditMenuGoBack)
}

func (dh deliveryHandler) handleMessages(c tele.Context) error {
	chatID := c.Chat().ID
	log := logmw.GetLogger(c)
	ctx := logr.NewContext(context.Background(), log)

	eventType, err := dh.taskUsecase.CurrentEventType(ctx, chatID)
	if err != nil {
		if errors.Is(err, sqlite_repo.ErrNoTaskEvent) {
			return c.Send(unknownAction, dh.mainMenu)
		}
		log.Error(err, "failed to get current event type")
		return c.Send(internalError)
	}
	switch eventType {
	case entities.TaskCreationEvent:
		return dh.handleCreationMessage(c, chatID)
	}
	return dh.handleEditMessage(c, chatID)
}

func (dh deliveryHandler) handleCreationMessage(c tele.Context, chatID int64) error {
	log := logmw.GetLogger(c)
	ctx := logr.NewContext(context.Background(), log)

	res, err := dh.taskUsecase.HandleTaskMessage(ctx, chatID, c.Message().Text)
	if err != nil {
		if errors.Is(err, sqlite_repo.ErrNoTaskEvent) {
			return c.Send(unknownAction, dh.mainMenu)
		} else if errors.Is(err, tasks.ErrParseRegularity) {
			return c.Send("Неверный формат регулярности напоминания, попробуйте еще раз")
		}
		log.Error(err, "failed to handle task message")
		return c.Send(internalError)
	}
	if res.IsTaskNameCreated() {
		return c.Send("Отлично! Как часто о ней надо напоминать?\n Ответь в формате N дней/недель/месяцев", dh.taskCreateStopMenu)
	}
	if res.IsTaskCreated() {
		chatTasks, err := dh.taskUsecase.GetTasks(ctx, chatID)
		if err != nil {
			log.Error(err, "failed to get tasks")
			return c.Send(internalError)
		}
		return c.Send("Прекрасно! Задачка создана, вы изумительны\n"+formatTasks(chatTasks), dh.mainMenu)
	}
	return c.Send("Я заблудился, напишите администратору @paulnopaul")
}

func (dh deliveryHandler) handleEditMessage(c tele.Context, chatID int64) error {
	res, err := dh.taskUsecase.HandleTaskMessage(context.Background(), chatID, c.Message().Text)
	if err != nil {
		if errors.Is(err, sqlite_repo.ErrNoTaskEvent) {
			return c.Send(unknownAction, dh.mainMenu)
		} else if errors.Is(err, tasks.ErrParseRegularity) {
			return c.Send("Неверный формат регулярности напоминания, попробуйте еще раз")
		} else if errors.Is(err, tasks.ErrBadTaskNumber) {
			return c.Send("Некорректный номер задачи, попробуйте еще раз")
		}
		log.Println(err)
		return c.Send(internalError)
	}
	if res.IsGotEditNumberTaskResult() {
		return c.Send("Выберите действие", dh.taskEditMenu)
	} else if res.IsGotEditNameTaskResult() {
		return c.Send("Название изменено, выберите действие", dh.taskEditMenu)
	} else if res.IsGotEditRegularityTaskResult() {
		return c.Send("Регулярность изменена, выберите действие", dh.taskEditMenu)
	}
	return c.Send("Я заблудился, напишите администратору @paulnopaul")
}

func (dh deliveryHandler) handleEditTaskName(c tele.Context) error {
	chatID := c.Chat().ID
	err := dh.taskUsecase.StartTaskNameEdit(context.Background(), chatID)
	if err != nil {
		if errors.Is(err, sqlite_repo.ErrNoTaskEvent) {
			return c.Send(unknownAction, dh.mainMenu)
		} else if errors.Is(err, tasks.ErrBadTaskEvent) {
			return c.Send("Вы не можете изменить имя, не начав редактировать задачу", dh.mainMenu)
		}
		return c.Send(internalError)
	}
	return c.Send("Как теперь будем ее называть?")
}

func (dh deliveryHandler) handleEditTaskRegularity(c tele.Context) error {
	chatID := c.Chat().ID
	err := dh.taskUsecase.StartTaskRegularityEdit(context.Background(), chatID)
	if err != nil {
		if errors.Is(err, sqlite_repo.ErrNoTaskEvent) {
			return c.Send(unknownAction, dh.mainMenu)
		} else if errors.Is(err, tasks.ErrBadTaskEvent) {
			return c.Send("Вы не можете изменить регулярность, не начав редактировать задачу", dh.mainMenu)
		}
		log.Println(err)
		return c.Send("Что-то пошло не так, почитай там логи что ли, лох")
	}
	return c.Send("Как теперь будем ее называть?")
}

func (dh deliveryHandler) handleEditGoBack(c tele.Context) error {
	chatID := c.Chat().ID
	ctx := context.Background()
	err := dh.taskUsecase.ResetTaskEdit(ctx, chatID)
	if err != nil {
		if errors.Is(err, sqlite_repo.ErrNoTaskEvent) {
			return c.Send(unknownAction, dh.mainMenu)
		} else if errors.Is(err, tasks.ErrBadTaskEvent) {
			return c.Send("Вы не можете это жмакнуть, не начав редактировать задачу", dh.mainMenu)
		}
		log.Println(err)
		return c.Send(internalError)
	}

	chatTasks, err := dh.taskUsecase.GetTasks(ctx, chatID)
	if err != nil {
		log.Println(err)
		return c.Send(internalError)
	}
	if len(chatTasks) == 0 {
		err := dh.taskUsecase.StopTaskEdit(ctx, chatID)
		if err != nil {
			log.Println(err)
			return c.Send(internalError)
		}
		return c.Send("У вас нет задач", dh.mainMenu)
	}
	return c.Send(formatTasks(chatTasks) + "Какую задачку будем менять? (введи номер)")
}

func (dh deliveryHandler) handleEditStop(c tele.Context) error {
	chatID := c.Chat().ID
	ctx := context.Background()
	err := dh.taskUsecase.StopTaskEdit(ctx, chatID)
	if err != nil {
		if errors.Is(err, sqlite_repo.ErrNoTaskEvent) {
			return c.Send(unknownAction, dh.mainMenu)
		} else if errors.Is(err, tasks.ErrBadTaskEvent) {
			return c.Send("Вы не можете это жмакнуть, не начав редактировать задачу", dh.mainMenu)
		}
		log.Println(err)
		return c.Send(internalError)
	}
	return c.Send("Что делать будем?", dh.mainMenu)
}

func (dh deliveryHandler) handleCreateStop(c tele.Context) error {
	chatID := c.Chat().ID
	log := logmw.GetLogger(c)
	ctx := logr.NewContext(context.Background(), log)
	err := dh.taskUsecase.StopTaskCreation(ctx, chatID)
	if err != nil {
		if errors.Is(err, sqlite_repo.ErrNoTaskEvent) {
			return c.Send(unknownAction, dh.mainMenu)
		} else if errors.Is(err, tasks.ErrBadTaskEvent) {
			return c.Send("Эту кнопку можно нажать только во время создания задачи", dh.mainMenu)
		}
		log.Error(err, "failed to stop task creation")
		return c.Send(internalError)
	}
	return c.Send("Что делать будем?", dh.mainMenu)
}

func (dh deliveryHandler) handleDeleteTask(c tele.Context) error {
	chatID := c.Chat().ID
	log := logmw.GetLogger(c)
	ctx := logr.NewContext(context.Background(), log)
	err := dh.taskUsecase.DeleteCurrentTask(ctx, chatID)
	if err != nil {
		if errors.Is(err, sqlite_repo.ErrNoTaskEvent) {
			return c.Send(unknownAction, dh.mainMenu)
		} else if errors.Is(err, tasks.ErrBadTaskEvent) {
			return c.Send("Вы не можете это жмакнуть, не начав редактировать задачу", dh.mainMenu)
		}
		log.Error(err, "failed to delete task")
		return c.Send(internalError)
	}
	chatTasks, err := dh.taskUsecase.GetTasks(ctx, chatID)
	if err != nil {
		log.Error(err, "cant get tasks")
		return c.Send(internalError)
	}
	if len(chatTasks) == 0 {
		return c.Send("У вас нет задач", dh.mainMenu)
	}
	return c.Send("Задача успешно удалена, ваши задачи:\n"+formatTasks(chatTasks), dh.mainMenu)
}

func formatTasks(tasks []entities.UserTask) string {
	res := "Ваши задачи:\n"
	for i, task := range tasks {
		// TODO: сделать красиво
		res += fmt.Sprintf("%d. %s каждые %d дней, до напоминания: %d дней\n", i+1, task.Name, int64(task.Regularity.Hours()/24), getRemindEst(task))
	}
	return res
}

func getRemindEst(task entities.UserTask) int64 {
	return int64(time.Until(task.LastReminded.Truncate(24*time.Hour).Add(task.Regularity).Add(task.RemindAfter)) / (24 * time.Hour))
}
