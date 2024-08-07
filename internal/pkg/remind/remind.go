package remind

import (
	"context"
	"fmt"
	"house-timer/internal/pkg/entities"
	"log"
	"time"

	tele "gopkg.in/telebot.v3"
)

type remindHanlder struct {
	taskRepo    entities.TaskStorage
	taskUsecase entities.TaskUsecase
	bot         *tele.Bot
	menu        *tele.ReplyMarkup
}

func NewRemindHandler(taskRepo entities.TaskStorage, taskUsecase entities.TaskUsecase, bot *tele.Bot) *remindHanlder {
	r := &remindHanlder{
		taskRepo:    taskRepo,
		taskUsecase: taskUsecase,
		bot:         bot,
	}

	remindMenu := &tele.ReplyMarkup{}
	btnTaskComplete := remindMenu.Data("Задача выполнена", "taskComplete")
	btnRemindAfter := remindMenu.Data("Напомнить позже", "remindAfter")

	remindMenu.Inline(
		remindMenu.Row(btnTaskComplete),
		remindMenu.Row(btnRemindAfter),
	)

	bot.Handle(&btnTaskComplete, r.handleTaskComplete)
	bot.Handle(&btnRemindAfter, r.handleRemindAfter)

	r.menu = remindMenu

	return r
}

func (r *remindHanlder) handleTaskComplete(c tele.Context) error {
	chatID := c.Chat().ID
	err := r.taskUsecase.CompleteTask(context.Background(), chatID)
	if err != nil {
		log.Println(err)
		return c.Send("Что-то пошло не так, почитай там логи что ли, лох")
	}
	return c.Send("Молодец огурец")
}

func (r *remindHanlder) handleRemindAfter(c tele.Context) error {
	chatID := c.Chat().ID
	err := r.taskUsecase.RemindLater(context.Background(), chatID)
	if err != nil {
		log.Println(err)
		return c.Send("Что-то пошло не так, почитай там логи что ли, лох")
	}
	return c.Send("Ок, напомню завтра")
}

func needsRemind(now time.Time, task entities.UserTask) bool {
	remindTime := task.LastReminded.Truncate(24 * time.Hour).Add(task.Regularity).Add(task.RemindAfter).Add(15 * time.Hour)
	return now.After(remindTime)
}

func (r *remindHanlder) remindTasks(ctx context.Context) {
	chats, err := r.taskRepo.GetChatIDs(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, chat := range chats {
		tasks, err := r.taskRepo.GetTasksForChat(ctx, chat)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(tasks)
		now := time.Now()
		for _, task := range tasks {
			if needsRemind(now, task) {
				res, err := r.taskUsecase.HandleRemind(ctx, task.ChatID)
				if err != nil {
					log.Print(err)
					continue
				}
				if res.IsNoRemindMessageResult() {
					r.bot.Send(&tele.User{ID: task.ChatID}, "Заканчивай, хочу напомнить тебе "+task.Name)
				} else if res.IsNeedRemindMessageResult() {
					r.bot.Send(&tele.User{ID: task.ChatID}, "Пора "+task.Name, r.menu)
				} else {
					log.Println("wtf remind res", res)
				}
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
