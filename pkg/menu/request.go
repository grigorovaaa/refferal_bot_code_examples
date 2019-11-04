package menu

import (
	"../context"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

type RequestMenu struct {
}

// Show отрисовываем меню
func (pm *RequestMenu) Show(ctx *context.Context) (err error) {
	menuLabels := []string{
		"Покупаю",
		"Продаю",
		"Назад",
	}

	text := "Выберите направление обмена:"

	row := tgbotapi.NewKeyboardButtonRow()
	keyboard := tgbotapi.ReplyKeyboardMarkup{ResizeKeyboard: true}
	for _, menuItem := range menuLabels {
		btn := tgbotapi.NewKeyboardButton(menuItem)
		row = append(row, btn)
	}
	keyboard.Keyboard = append(keyboard.Keyboard, row)

	var msg tgbotapi.MessageConfig
	msg = tgbotapi.NewMessage(ctx.ChatId, text)
	msg.ReplyMarkup = keyboard

	if _, err = ctx.Dao.Bot.Send(msg); err != nil {
		return
	}

	return nil
}

// Update обрабатываем поведение клиента в меню
func (pm *RequestMenu) Update(ctx *context.Context, u *tgbotapi.Update) (err error) {
	switch u.Message.Text {
	case "Покупаю":
		if err = ctx.NextMenu(ctx.Dao.AllMenus["*menu.RequestCurrencyFromMenu"]); err != nil {
			return
		}

	case "Продаю":
		if err = ctx.NextMenu(ctx.Dao.AllMenus["*menu.RequestCurrencyToMenu"]); err != nil {
			return
		}

	case "Назад":
		if err = ctx.PrevMenu(); err != nil {
			return
		}

	default:
		text := "Неизвестная команда"
		msg := tgbotapi.NewMessage(ctx.ChatId, text)

		if _, err = ctx.Dao.Bot.Send(msg); err != nil {
			return
		}
	}
	return nil
}
