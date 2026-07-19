package aiclient

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter interface {
	Wait(ctx context.Context) error
}
type ConnRedis struct {
	redisClient *redis.Client
	key         string
	period      time.Duration
	limit       int
}

func NewConnRedis(rdb *redis.Client, key string, period time.Duration, limit int) *ConnRedis {
	return &ConnRedis{redisClient: rdb, key: key, period: period, limit: limit}
}

func (c *ConnRedis) Wait(ctx context.Context) error {
	for {
		rdb := c.redisClient

		val, err := rdb.Incr(ctx, c.key).Result()
		if err != nil {
			log.Println(err)
			return err
		}

		// val == 1 - первый запрос в окне, выставляем TTL.
		if val == 1 {
			_, err = rdb.Expire(ctx, c.key, c.period).Result()
			if err != nil {
				log.Println(err)
				return err
			}
		}

		// лимит превышен — ждём истечения TTL и пробуем снова.
		if val > int64(c.limit) {
			ttl, err := rdb.TTL(ctx, c.key).Result()
			if err != nil {
				log.Println(err)
				return err
			}
			select {
			case <-time.After(ttl):
				//Если время превышено, ttl умирает,
				//программа переходит к выполненному кейсу
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	}
}
