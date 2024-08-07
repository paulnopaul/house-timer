package delivery

import (
	"context"
	"errors"
	"fmt"
	"house-timer/internal/pkg/repos/sqlite_repo"
	"house-timer/internal/pkg/usecases/tasks"
	"log"

	"house-timer/internal/pkg/entities"

	"github.com/go-logr/logr"
	tele "gopkg.in/telebot.v3"
)

type deliveryHandler struct {
	mainMenu           *tele.ReplyMarkup
	taskEditMenu       *tele.ReplyMarkup
	taskEditMenuGoBack *tele.ReplyMarkup
	logger             logr.Logger

	taskUsecase entities.TaskUsecase
}

const loggerKey = "logger"

func (dh *deliveryHandler) Logger(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		c.Set(loggerKey, dh.logger.WithValues("chatID", c.Chat().ID, "reqID", c.Message().ID))
		return next(c)
	}
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
	btnEditStop2 := taskEditMenuGoBack.Data("Закончить изменение задач", "taskEditStop2")
	taskEditMenu.Inline(
		taskEditMenu.Row(btnEditStop2),
	)

	dh := deliveryHandler{
		mainMenu:           mainMenu,
		taskEditMenu:       taskEditMenu,
		taskEditMenuGoBack: taskEditMenuGoBack,
		logger:             logr.Logger{}.V(4),

		taskUsecase: taskUsecase,
	}

	bot.Use(dh.Logger)
	bot.Handle("/start", dh.handleStart)
	bot.Handle(&btnNewTask, dh.handleNewTask)
	bot.Handle(&btnEditTask, dh.handleEditTask)

	bot.Handle(&btnEditName, dh.handleEditTaskName)
	bot.Handle(&btnEditRegularity, dh.handleEditTaskRegularity)
	bot.Handle(&btnDeleteTask, dh.handleDeleteTask)

	bot.Handle(&btnEditGoBack, dh.handleEditGoBack)
	bot.Handle(&btnEditStop, dh.handleEditStop)
	bot.Handle(&btnEditStop2, dh.handleEditStop)

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

	log := c.Get(loggerKey).(logr.Logger)
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
	return c.Send("О чем надо напоминать?")
}

func (dh deliveryHandler) handleEditTask(c tele.Context) error {
	chatID := c.Chat().ID
	log := c.Get(loggerKey).(logr.Logger)
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
	log := c.Get(loggerKey).(logr.Logger)
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
	log := c.Get(loggerKey).(logr.Logger)
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
		return c.Send("Отлично! Как часто о ней надо напоминать?\n Ответь в формате N дней/недель/месяцев")
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

func (dh deliveryHandler) handleDeleteTask(c tele.Context) error {
	chatID := c.Chat().ID
	log := c.Get(loggerKey).(logr.Logger)
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
	return c.Send("Что делать будем?", dh.mainMenu)
}

func formatTasks(tasks []entities.UserTask) string {
	res := "Ваши задачи:\n"
	for i, task := range tasks {
		// TODO: сделать красиво
		res += fmt.Sprintf("%d. %s каждые %d дней\n", i+1, task.Name, int64(task.Regularity.Hours()/24))
	}
	return res
}
