package context

import (
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

const (
	redisCtxMenuKey = "referralBot:ctx:menu"
	redisCtxMenuTtl = 86400 // ttl жизни истории пройденных меню
)

// setMenu сохранить историю пройденных меню
func setMenu(ctx *Context, encodeCtxMenus string) (err error) {
	redisKey := redisCtxMenuKey + ":" + ctx.Username + ":" + fmt.Sprintf("%d", ctx.ChatId)
	ttl := time.Duration(redisCtxMenuTtl) * time.Second

	if err = ctx.Dao.RedisClient.Set(redisKey, encodeCtxMenus, ttl).Err(); err != nil {
		return
	}
	return
}

// getMenu получить историю пройденных меню
func getMenu(ctx *Context) (encodeCtxMenus string, err error) {
	redisKey := redisCtxMenuKey + ":" + ctx.Username + ":" + fmt.Sprintf("%d", ctx.ChatId)

	if encodeCtxMenus, err = ctx.Dao.RedisClient.Get(redisKey).Result(); err != nil {
		if err != redis.Nil {
			return
		} else {
			err = nil
		}
	}
	return
}
