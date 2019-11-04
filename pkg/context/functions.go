package context

// Load загрузить контекст клиента
func Load(dao *DAO, chatId int64, username string) (ctx *Context, err error) {
	// чистим контекст
	ctx = &Context{}
	ctx.Dao = dao
	ctx.Ex = &ExchangeCtx{}
	ctx.Menus = make([]Menu, 0)
	ctx.ChatId = chatId
	ctx.Username = username

	// достать сохраненные переходы по меню из репозитория
	var encodeCtxMenus string
	if encodeCtxMenus, err = getMenu(ctx); err != nil {
		return
	}

	// достать сохраненную заявку на обмен из репозитория
	var encodeCtxExchange string
	if encodeCtxExchange, err = getExchange(ctx); err != nil {
		return
	}

	// загрузим то, что достали из репозиториев в контекст
	if err = ctx.UnSerialize(encodeCtxMenus, encodeCtxExchange); err != nil {
		return
	}

	return
}

// Save сохранить контекст клиента
func Save(ctx *Context) (err error) {
	// сериализуем контекст в вид, пригодный для репозиториев
	var encodeCtxMenus string
	var encodeCtxExchange string
	if encodeCtxMenus, encodeCtxExchange, err = ctx.Serialize(); err != nil {
		return
	}

	// сохраним заявку на обмен в репозиторий
	if err = setMenu(ctx, encodeCtxMenus); err != nil {
		return
	}

	// сохраним переходы по меню в репозиторий
	if err = setExchange(ctx, encodeCtxExchange); err != nil {
		return
	}

	return
}
