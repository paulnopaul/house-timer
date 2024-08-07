package remind

import (
	"context"
	"house-timer/internal/pkg/entities"
	"house-timer/internal/pkg/logmw"
	"log"
	"log/slog"
	"time"

	"github.com/go-logr/logr"
	tele "gopkg.in/telebot.v3"
)

type remindHanlder struct {
	taskRepo    entities.TaskStorage
	taskUsecase entities.TaskUsecase
	bot         *tele.Bot
	menu        *tele.ReplyMarkup
	logger      logr.Logger
}

func NewRemindHandler(taskRepo entities.TaskStorage, taskUsecase entities.TaskUsecase, bot *tele.Bot) *remindHanlder {
	r := &remindHanlder{
		taskRepo:    taskRepo,
		taskUsecase: taskUsecase,
		bot:         bot,
		logger:      logr.FromSlogHandler(slog.NewTextHandler(log.Writer(), nil)),
	}

	remindMenu := &tele.ReplyMarkup{}
	btnTaskComplete := remindMenu.Data("Задача выполнена", "taskComplete")
	btnRemindAfter := remindMenu.Data("Напомнить позже", "remindAfter")

	remindMenu.Inline(
		remindMenu.Row(btnTaskComplete),
		remindMenu.Row(btnRemindAfter),
	)

	bot.Use(logmw.NewLogMW(r.logger))
	bot.Handle(&btnTaskComplete, r.handleTaskComplete)
	bot.Handle(&btnRemindAfter, r.handleRemindAfter)

	r.menu = remindMenu

	return r
}

func (r *remindHanlder) handleTaskComplete(c tele.Context) error {
	chatID := c.Chat().ID
	log := logmw.GetLogger(c)
	err := r.taskUsecase.CompleteTask(context.Background(), chatID)
	if err != nil {
		log.Error(err, "failed to complete task")
		return c.Send("Что-то пошло не так, почитай там логи что ли, лох")
	}
	return c.Send("Молодец огурец")
}

func (r *remindHanlder) handleRemindAfter(c tele.Context) error {
	chatID := c.Chat().ID
	log := logmw.GetLogger(c)
	err := r.taskUsecase.RemindLater(context.Background(), chatID)
	if err != nil {
		log.Error(err, "failed to remind later")
		return c.Send("Что-то пошло не так, почитай там логи что ли, лох")
	}
	return c.Send("Ок, напомню завтра")
}

func needsRemind(now time.Time, task entities.UserTask) bool {
	remindTime := task.LastReminded.Truncate(24 * time.Hour).Add(task.Regularity).Add(task.RemindAfter).Add(15 * time.Hour)
	return now.After(remindTime)
}

func (r *remindHanlder) remindTasks(ctx context.Context) {
	log := r.logger.WithName("remind loop").WithValues("id", time.Now().Unix())
	chats, err := r.taskRepo.GetChatIDs(ctx)
	if err != nil {
		log.Error(err, "failed to get chat ids")
		return
	}
	for _, chat := range chats {
		log.Info("reminding tasks for chat", "chat", chat)
		tasks, err := r.taskRepo.GetTasksForChat(ctx, chat)
		if err != nil {
			log.Error(err, "failed to get tasks for chat", "chat", chat)
			return
		}
		now := time.Now()
		for _, task := range tasks {
			if needsRemind(now, task) {
				res, err := r.taskUsecase.HandleRemind(ctx, task.ChatID)
				if err != nil {
					log.Error(err, "failed to handle remind", "task", task)
					continue
				}
				if res.IsNoRemindMessageResult() {
					_, err = r.bot.Send(&tele.User{ID: task.ChatID}, "Заканчивай, хочу напомнить тебе "+task.Name)
					if err != nil {
						log.Error(err, "failed to send remind message", "taskID", task.ID)
					}
				} else if res.IsNeedRemindMessageResult() {
					_, err := r.bot.Send(&tele.User{ID: task.ChatID}, "Пора "+task.Name, r.menu)
					if err != nil {
						log.Error(err, "failed to send remind message with menu", "taskID", task.ID)
					}
				} else {
					log.Error(nil, "unexpected remind result", "result", res)
				}
				log.Info("reminded task", "taskID", task.ID)
			}
		}
	}
}

func (r *remindHanlder) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Minute):
			r.remindTasks(ctx)
		}
	}
}
