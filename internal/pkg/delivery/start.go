package delivery

import (
	"context"
	"fmt"
	"house-timer/internal/pkg/entities"
	"log"

	tele "gopkg.in/telebot.v3"
)

type deliveryHandler struct {
	mainMenu *tele.ReplyMarkup

	taskUsecase entities.TaskUsecase
}

func NewDeliveryHandler(bot *tele.Bot, taskUsecase entities.TaskUsecase) {

	selector := &tele.ReplyMarkup{}
	btnNewTask := selector.Data("Создад", "createTask")
	selector.Inline(
		selector.Row(btnNewTask),
	)

	dh := deliveryHandler{
		mainMenu:    selector,
		taskUsecase: taskUsecase,
	}
	bot.Handle("/start", dh.handleStart)
	bot.Handle(&btnNewTask, dh.handleNewTask)

	bot.Handle(tele.OnText, dh.handleMessages)
}

func (dh deliveryHandler) handleStart(c tele.Context) error {
	chatID := c.Chat().ID
	return c.Send(fmt.Sprint("Превед я бот зодачнегк в чате намбер", chatID), dh.mainMenu)
}

func (dh deliveryHandler) handleNewTask(c tele.Context) error {
	chatID := c.Chat().ID

	err := dh.taskUsecase.CreateEmptyTask(context.Background(), chatID)
	if err != nil {
		log.Println(err)
		return c.Send("Что-то пошло не так, почитай там логи что ли, лох")
	}
	return c.Send("довай дадим имя твоей задачке\nНапиши имя как хош")
}

func (dh deliveryHandler) handleMessages(c tele.Context) error {
	chatID := c.Chat().ID
	res, err := dh.taskUsecase.HandleTaskMessage(context.Background(), chatID, c.Message().Text)
	if err != nil {
		log.Println(err)
		return c.Send("Что-то пошло не так, почитай там логи что ли, лох")
	}
	if res.IsTaskNameCreated() {
		return c.Send("Отлично! Как часто о ней надо напоминать?\n Ответь в формате N дней/недель/месяцев")
	}
	if res.IsTaskCreated() {
		tasks, err := dh.taskUsecase.GetTasks(context.Background(), chatID)
		if err != nil {
			log.Println(err)
			return c.Send("Что-то пошло не так, почитай там логи что ли, лох")
		}
		return c.Send("Прекрасно! Задачка создана, вы изумительны\n"+formatTasks(tasks), dh.mainMenu)
	}
	return c.Send("Я заблудился, напишите администратору @paulnopaul")
}

func formatTasks(tasks []entities.UserTask) string {
	res := "Ваши задачи:\n"
	for i, task := range tasks {
		// TODO: сделать красиво
		res += fmt.Sprintf("%d. %s каждые %d дней\n", i+1, task.Name, int64(task.Regularity.Hours() / 24))
	}
	return res
}

// func createMenu(b *tele.Bot) {
// 	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
// 	selector := &tele.ReplyMarkup{}

// 	btnHelp := menu.Text("ℹ Help")
// 	btnSettings := menu.Text("⚙ Settings")

// 	btnPrev := selector.Data("⬅", "prev")
// 	btnNext := selector.Data("➡", "next")
// 	menu.Reply(
// 		menu.Row(btnHelp),
// 		menu.Row(btnSettings),
// 	)

// 	selector.Inline(
// 		selector.Row(btnPrev, btnNext),
// 	)

// 	b.Handle("/start", func(c tele.Context) error {
// 		return c.Send("Hello!", selector)
// 	})

// 	// On reply button pressed (message)
// 	b.Handle(&btnHelp, func(c tele.Context) error {
// 		return c.Edit("Here is some help: ...")
// 	})

// 	// On inline button pressed (callback)
// 	b.Handle(&btnPrev, func(c tele.Context) error {
// 		return c.Respond()
// 	})

// }
