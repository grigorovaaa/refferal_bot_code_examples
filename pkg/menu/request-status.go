package menu

import (
	"../context"
	"../domain/exchange-request"
	bot_chat "../telegram-bot"
	"../utils"
	"fmt"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"strings"
)

type RequestStatusMenu struct {
}

// Show отрисовываем меню
func (pm *RequestStatusMenu) Show(ctx *context.Context) (err error) {
	menuLabels := []string{
		"Да",
		"Нет",
	}

	amountIn, amountOut := strAmounts(ctx)

	text := fmt.Sprintf("*Получаете*: %s%s.\n*Отдаете*: %s%s.\n*Способ*: %s.\nВерно?",
		amountIn, ctx.Ex.CurrencyFrom, amountOut, ctx.Ex.CurrencyTo, ctx.Ex.CashWay)

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
	msg.ParseMode = "markdown"

	if _, err = ctx.Dao.Bot.Send(msg); err != nil {
		return
	}

	return nil
}

// Update обрабатываем поведение клиента в меню
func (pm *RequestStatusMenu) Update(ctx *context.Context, u *tgbotapi.Update) (err error) {
	switch u.Message.Text {
	case "Да":
		// сохранить в бд подтверхжение
		if ctx.Ex.Id, err = utils.GenerateId(utils.LenId); err != nil {
			return
		}
		if err = exchange_request.Save(ctx); err != nil {
			return
		}

		amountIn, amountOut := strAmounts(ctx)

		text := "Благодарим за обращение!\nВаш запрос принят.\n\nВ ближайшее время с Вами свяжется оператор для выставления предложения.\n\nПожалуйста, сохраните номер Вашего запроса"
		msg := tgbotapi.NewMessage(ctx.ChatId, text)
		msg.ParseMode = "markdown"
		if _, err = ctx.Dao.Bot.Send(msg); err != nil {
			return
		}

		text = "Номер запроса: %s"
		text = fmt.Sprintf(text, ctx.Ex.Id)
		msg = tgbotapi.NewMessage(ctx.ChatId, text)
		if _, err = ctx.Dao.Bot.Send(msg); err != nil {
			return
		}

		// прокинуть в чат менеджерам подтверждение
		escapedUsername := strings.ReplaceAll(ctx.Username, "_", "\\_")
		botMsg := fmt.Sprintf(bot_chat.GetMessage("accept_request"), ctx.Ex.TypeLabel()+"у", escapedUsername, ctx.Ex.Id, amountIn, ctx.Ex.CurrencyFrom, amountOut, ctx.Ex.CurrencyTo, ctx.Ex.CashWay)
		bot_chat.Message(botMsg, ctx.Dao.Log)

		// отправить на стартовое меню
		if err = ctx.WellComeMenu(); err != nil {
			return
		}

	case "Нет":
		// удалить из временного хранилища запрос на обмен
		ctx.Ex = &context.ExchangeCtx{}

		// отправить на стартовое меню
		if err = ctx.WellComeMenu(); err != nil {
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

func strAmounts(ctx *context.Context) (amountIn, amountOut string) {
	if ctx.Ex.AmountIn != 0 {
		amountIn = fmt.Sprintf("%0.8g ", ctx.Ex.AmountIn)
	} else {
		amountIn = ""
	}
	if ctx.Ex.AmountOut != 0 {
		amountOut = fmt.Sprintf("%0.8g ", ctx.Ex.AmountOut)
	} else {
		amountOut = ""
	}

	return
}
