package menu

import (
	"../context"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"strconv"
)

type RequestAmountInMenu struct {
}

// Show отрисовываем меню
func (pm *RequestAmountInMenu) Show(ctx *context.Context) (err error) {
	menuLabels := []string{
		"Назад",
	}

	text := "Введите сумму покупки (разделитель точка)"

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
func (pm *RequestAmountInMenu) Update(ctx *context.Context, u *tgbotapi.Update) (err error) {
	ctx.Ex.AmountIn = 0

	switch u.Message.Text {
	case "Назад":
		if err = ctx.PrevMenu(); err != nil {
			return
		}
	default:
		// проверим, ввели число и положительное
		var amount float64
		if amount, err = strconv.ParseFloat(u.Message.Text, 64); err != nil {
			text := "Неправильный формат суммы. Введите сумму покупки (разделитель точка) или нажмите кнопку \"Назад\""
			msg := tgbotapi.NewMessage(ctx.ChatId, text)

			if _, err = ctx.Dao.Bot.Send(msg); err != nil {
				return
			}
			// возвращаемся потому, что невалидное ввели
			return
		}

		if amount <= 0 {
			text := "Сумма должна быть больше 0"
			msg := tgbotapi.NewMessage(ctx.ChatId, text)

			if _, err = ctx.Dao.Bot.Send(msg); err != nil {
				return
			}
			// возвращаемся потому, что невалидное ввели
			return
		}

		ctx.Ex.AmountIn = amount

		if err = ctx.NextMenu(ctx.Dao.AllMenus["*menu.RequestCurrencyToMenu"]); err != nil {
			return
		}
	}

	return nil
}
