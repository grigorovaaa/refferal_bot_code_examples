package context

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"reflect"
)

type DAO struct {
	Bot         *tgbotapi.BotAPI
	Log         *logrus.Entry
	Conn        *sql.DB
	RedisClient *redis.Client
	AllMenus    map[string]Menu
}

type ExchangeCtx struct {
	Id           string
	CurrencyFrom string
	CurrencyTo   string
	AmountIn     float64
	AmountOut    float64
	CashWay      string
}

func (ex *ExchangeCtx) TypeLabel() string {
	if ex.AmountIn != 0 {
		return "покупк"
	} else if ex.AmountOut != 0 {
		return "продаж"
	} else {
		return ""
	}
}

type Menu interface {
	Show(ctx *Context) error
	Update(ctx *Context, u *tgbotapi.Update) error
}

type Context struct {
	ChatId   int64
	Username string
	Menus    []Menu
	Ex       *ExchangeCtx
	Dao      *DAO
}

// Update обработка действия из меню
func (ctx *Context) Update(u *tgbotapi.Update) error {
	// тут оно отправляет Update в последний элемент в массиве ctx.Menus
	i := len(ctx.Menus)
	// если есть меню - переход
	if i > 0 {
		m := ctx.Menus[i-1]
		return m.Update(ctx, u)
		// если нет ни одного - отобразить стартовое
	} else {
		return ctx.WellComeMenu()
	}
}

// WellComeMenu переход на стартовое меню
func (ctx *Context) WellComeMenu() error {
	// бнулить все, что накопила раньше - и запрос и историю меню
	ctx.Ex = &ExchangeCtx{}
	ctx.Menus = make([]Menu, 0)

	return ctx.NextMenu(ctx.Dao.AllMenus["*menu.WelcomeMenu"])
}

// NextMenu переход на следющее меню
func (ctx *Context) NextMenu(menu Menu) error {
	ctx.Menus = append(ctx.Menus, menu)
	return menu.Show(ctx)
}

// PrevMenu переход на предыдущее меню
func (ctx *Context) PrevMenu() error {
	// удалить последниее меню
	i := len(ctx.Menus)
	ctx.Menus = ctx.Menus[:i-1]
	// отобразит предпоследнее меню
	return ctx.Menus[i-2].Show(ctx)
}

// Serialize сериализация контекста
func (ctx *Context) Serialize() (encodeCtxMenus, encodeCtxExchange string, err error) {
	var bytes []byte

	// menu
	arr := make([]string, 0)
	for _, currentMenu := range ctx.Menus {
		arr = append(arr, reflect.TypeOf(currentMenu).String())
	}

	if bytes, err = json.Marshal(arr); err != nil {
		return
	}
	encodeCtxMenus = string(bytes)

	// exchange
	if bytes, err = json.Marshal(ctx.Ex); err != nil {
		return
	}

	encodeCtxExchange = string(bytes)

	return
}

// UnSerialize десериализация контекста
func (ctx *Context) UnSerialize(encodeCtxMenus, encodeCtxExchange string) (err error) {
	var bytes []byte
	// menu
	if encodeCtxMenus != "" {
		menuArr := make([]string, 0)
		bytes = []byte(encodeCtxMenus)
		if err = json.Unmarshal(bytes, &menuArr); err != nil {
			return
		}

		var menuType Menu
		var ok bool
		for _, menuLabel := range menuArr {

			if menuType, ok = ctx.Dao.AllMenus[menuLabel]; !ok {
				err = errors.New(fmt.Sprintf("неизвестное меню: %s", menuLabel))
				return
			}

			ctx.Menus = append(ctx.Menus, menuType)
		}

		if encodeCtxExchange != "" {
			// exchange
			bytes = []byte(encodeCtxExchange)
			if err = json.Unmarshal(bytes, &ctx.Ex); err != nil {
				return
			}
		}
	}

	return
}
