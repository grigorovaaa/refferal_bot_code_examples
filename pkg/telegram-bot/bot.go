package telegram_bot

import (
	"../config"
	"fmt"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

// todo если в собщении бота есть спецсимфволы из маркдауна * _ которые не про разметку, а про сообщение. то все поломается и не придет письмо
// 	такое письмо НЕ приходит message = "*Новое сообщение* тема обращения: subject сообщение: _text \n"
// 	такое письмо приходит message = "*Новое сообщение* тема обращения: subject сообщение: \\_text \n"
var messages = map[string]string{
	"accept_request": "*Запрос на %s* \n пользователь: @%s \n номер: %s \n Получает: %s%s; Отдает %s%s; Способ: %s",
}

// send отправляет сообщение от конкретного бота в конкретный канал
func send(telegramBotToken string, chatId int64, msg string, log *logrus.Entry) {
	var err error
	// получить бота по токену
	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Error(err)
		return
	}

	// отправить боту сообщение
	msg = fmt.Sprintf(msg)
	tlgMsg := tgbotapi.NewMessage(chatId, msg)
	tlgMsg.ParseMode = "markdown"
	_, err = bot.Send(tlgMsg)
	if err != nil {
		log.Error(err)
		return
	}
}

// GetMessage возвращает сообщение по ключу из карты (карта невидимая миру, а тут внутри никто ее не меняет)
func GetMessage(key string) (message string) {
	var ok bool
	if message, ok = messages[key]; !ok {
		message = "Произошло какое-то событие, тип которого не определен. разбираюсь..."
	}
	return
}

// Message отправляет сообщение в канал для бизнесса (про бизнесс события тут речь)
// todo логгер тут чтобы о проблемах бота в лог сообщать можно было - ему глобальным бы быть?
func Message(msg string, log *logrus.Entry) {
	// получить токен бота - если нет в настройках его, считаем что бот не нужен приложению
	telegramBotToken := config.GetInstance().Bot.Token
	if telegramBotToken == "" {
		return
	}

	// получить чат для сообщения - если нет в настройках его, считаем что данный чат не нужен приложению
	chatId := config.GetInstance().Bot.NoticeChat
	if chatId == 0 {
		return
	}

	send(telegramBotToken, chatId, msg, log)

	return
}

// Error отправляет сообщение об ошибках в канал для разработчиков
func Error(msg string, log *logrus.Entry) {
	// получить токен бота - если нет в настройках его, считаем что бот не нужен приложению
	telegramBotToken := config.GetInstance().ServiceBot.Token
	if telegramBotToken == "" {
		return
	}

	// получить чат для сообщения - если нет в настройках его, считаем что данный чат не нужен приложению
	chatId := config.GetInstance().ServiceBot.ErrorChat
	if chatId == 0 {
		return
	}

	send(telegramBotToken, chatId, msg, log)

	return
}

// Info отправляет информационные сообщение в канал для разработчиков (про то, что кроны запустились и отработали тут)
func Info(msg string, log *logrus.Entry) {
	// получить токен бота - если нет в настройках его, считаем что бот не нужен приложению
	telegramBotToken := config.GetInstance().ServiceBot.Token
	if telegramBotToken == "" {
		return
	}

	// получить чат для сообщения - если нет в настройках его, считаем что данный чат не нужен приложению
	chatId := config.GetInstance().ServiceBot.InfoChat
	if chatId == 0 {
		return
	}

	send(telegramBotToken, chatId, msg, log)

	return
}
