package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

func NewRedisDB(ctx context.Context, host, port, password string) *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       0,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		panic("Unable to connect to redis " + err.Error())
	}
	return redisClient
}
