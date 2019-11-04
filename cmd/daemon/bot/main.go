package main

import (
	"../../../pkg/config"
	"../../../pkg/context"
	"../../../pkg/domain/bot-session"
	"../../../pkg/logger"
	"../../../pkg/menu"
	"../../../pkg/utils"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"runtime"
	"strconv"
)

var (
	dao *context.DAO
)

const (
	redisDB = 1
)

func init() {
	var err error
	// парсим аргументы
	confPathPtr := flag.String("conf", "", "[string] absolute path to configuration file")
	flag.Parse()

	_log := logger.New()
	_log.Infof("GOMAXPROCS %d", runtime.NumCPU())

	// подгружаем конфиг
	conf := config.Config{}
	err = conf.LoadFromFile(*confPathPtr)
	if err != nil {
		_log.Panic(err)
	}
	conf.SaveInstance()

	// настраиваем логирование
	var log *logrus.Entry
	log = _log.WithFields(logrus.Fields{
		"db_host": conf.DB.Host,
		"db_port": conf.DB.Port,
		"db_name": conf.DB.Name,
		"bot":     conf.Bot.Token,
	})

	// конект к базе
	dbUrl := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", conf.DB.Host, conf.DB.Port, conf.DB.User, conf.DB.Password, conf.DB.Name)
	var conn *sql.DB
	if conn, err = sql.Open("postgres", dbUrl); err != nil {
		utils.CatchPanic(log, err)
	}

	// -- redis client
	var redisClient *redis.Client
	redisClient = redis.NewClient(&redis.Options{
		Addr:     config.GetInstance().Redis.Host + ":" + strconv.Itoa(config.GetInstance().Redis.Port),
		Password: config.GetInstance().Redis.Password,
		DB:       redisDB,
	})

	// токен бота из конфига
	telegramChattingBotToken := config.GetInstance().Bot.Token
	// без него не запускаемся
	if telegramChattingBotToken == "" {
		err = errors.New("-telegram bot token is required")
		utils.CatchPanic(log, err)
	}
	// используя токен создаем новый инстанс бота
	var bot *tgbotapi.BotAPI
	if bot, err = tgbotapi.NewBotAPI(telegramChattingBotToken); err != nil {
		utils.CatchPanic(log, err)
	}

	bot, err = tgbotapi.NewBotAPI(telegramChattingBotToken)
	if err != nil {
		utils.CatchPanic(log, err)
	}

	// карта со всеми меню
	allMenus := getAllMenus()

	// объекты доступа к данным и прочие общие вещи
	dao = &context.DAO{
		Bot:         bot,
		Log:         log,
		Conn:        conn,
		RedisClient: redisClient,
		AllMenus:    allMenus,
	}
}

// getAllMenus формирует объекты всех меню - они всегда в памяти
// и их набираем по ходу для каждого клиента, а не формируем заново
func getAllMenus() (menusMap map[string]context.Menu) {

	menusMap = map[string]context.Menu{
		"*menu.WelcomeMenu":             &menu.WelcomeMenu{},
		"*menu.AboutMenu":               &menu.AboutMenu{},
		"*menu.RequestMenu":             &menu.RequestMenu{},
		"*menu.RequestCurrencyFromMenu": &menu.RequestCurrencyFromMenu{},
		"*menu.RequestCurrencyToMenu":   &menu.RequestCurrencyToMenu{},
		"*menu.RequestAmountInMenu":     &menu.RequestAmountInMenu{},
		"*menu.RequestAmountOutMenu":    &menu.RequestAmountOutMenu{},
		"*menu.RequestCashWayMenu":      &menu.RequestCashWayMenu{},
		"*menu.RequestStatusMenu":       &menu.RequestStatusMenu{},
	}
	return
}

func main() {
	// u - структура с конфигом для получения апдейтов
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// используя конфиг u создаем канал в который будут прилетать новые сообщения
	updates, err := dao.Bot.GetUpdatesChan(u)

	// переменная контекста
	var ctx *context.Context

	// в канал updates прилетают структуры типа Update вычитываем их и обрабатываем
	for update := range updates {
		var allow bool
		if allow, err = checkUpdate(&update); err != nil {
			utils.CatchErr(dao.Log, err)
			continue
		}
		if allow == false {
			continue
		}

		// загрузить контекст текущего пользователя
		if ctx, err = context.Load(dao, update.Message.Chat.ID, update.Message.Chat.UserName); err != nil {
			utils.CatchErr(dao.Log, err)
			continue
		}

		// пришло сообщение - формирует ответ в зависимости от того, что нам написали
		if update.Message != nil {

			// todo сделать логирование в текстовый файл
			// логируем от кого какое сообщение пришло
			ctx.Dao.Log.Printf("[%s] [%s] %s", update.Message.From.UserName, update.Message.From.ID, update.Message.Text)

			// обработка комманд (комманда - сообщение, начинающееся с "/")
			if update.Message.Command() != "" {
				switch update.Message.Command() {
				case "start":

					if err = ctx.WellComeMenu(); err != nil {
						utils.CatchErr(dao.Log, err)
						continue
					}

					// сохраним контекст текущего пользователя
					if err = context.Save(ctx); err != nil {
						utils.CatchErr(dao.Log, err)
						continue
					}

					// сохранить пользователя в БД постгреса - нигде не используется, но можно статистику будет достать
					if err = bot_session.Save(ctx.Dao.Conn, update.Message.Chat.UserName, update.Message.From.ID, update.Message.Chat.ID); err != nil {
						utils.CatchErr(dao.Log, err)
						continue
					}

				default:
					text := "Неизвестная команда"
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
					if _, err = ctx.Dao.Bot.Send(msg); err != nil {
						utils.CatchErr(dao.Log, err)
						continue
					}
					continue
				}

				// обработка сообщений (нажатия меню приравниваются к сообщениям)
			} else {
				if err = ctx.Update(&update); err != nil {
					utils.CatchErr(dao.Log, err)
					continue
				}

				// сохраним контекст текущего пользователя
				if err = context.Save(ctx); err != nil {
					utils.CatchErr(dao.Log, err)
					continue
				}
			}
		}
	}
}

// checkUpdate проверяет сообщение, имеет ли смысл на него отвечать
func checkUpdate(update *tgbotapi.Update) (allow bool, err error) {
	allow = true
	// пользователь без username - игнорируем таких
	if update.Message.Chat.UserName == "" {
		text := "У Вас в телеграмме не установлен username - мы не сможем с Вами связаться. Если хотите продолжить работу, то установите его (Настройки->Изменить профиль / Settings->Edit profile)"
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
		if _, err = dao.Bot.Send(msg); err != nil {
			return
		}

		// аких без имени тоже сохраним в БД - для статистики хотя бы, сколько их было таких
		if err = bot_session.Save(dao.Conn, update.Message.Chat.UserName, update.Message.From.ID, update.Message.Chat.ID); err != nil {
			return
		}

		allow = false
	}

	return
}
