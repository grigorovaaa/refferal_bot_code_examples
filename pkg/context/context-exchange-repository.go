package context

import (
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

const (
	redisCtxExKey = "referralBot:ctx:exchange"
	redisCtxExTtl = 86400 // ttl жизни формируемого запроса для заявки на обмен денег
)

// setExchange сохранить формируемый запрос для заявки на обмен денег
func setExchange(ctx *Context, encodeCtxMenus string) (err error) {
	redisKey := redisCtxExKey + ":" + ctx.Username + ":" + fmt.Sprintf("%d", ctx.ChatId)
	ttl := time.Duration(redisCtxExTtl) * time.Second

	if err = ctx.Dao.RedisClient.Set(redisKey, encodeCtxMenus, ttl).Err(); err != nil {
		return
	}
	return
}

// getExchange получить формируемый запрос для заявки на обмен женег
func getExchange(ctx *Context) (encodeCtxMenus string, err error) {
	redisKey := redisCtxExKey + ":" + ctx.Username + ":" + fmt.Sprintf("%d", ctx.ChatId)

	if encodeCtxMenus, err = ctx.Dao.RedisClient.Get(redisKey).Result(); err != nil {
		if err != redis.Nil {
			return
		} else {
			err = nil
		}
	}
	return
}
