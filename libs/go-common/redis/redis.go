
package redis

import (
	"context"
	"fmt"
	"log"

	"github.com/generator-platform/go-common/config"
	"github.com/redis/go-redis/v9"
)

var Client *redis.Client
var ctx = context.Background()

func InitRedis(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	Client = client
	log.Println("Redis connected successfully")
	return client, nil
}

func Set(key string, value interface{}, expiration int) error {
	return Client.Set(ctx, key, value, 0).Err()
}

func Get(key string) (string, error) {
	return Client.Get(ctx, key).Result()
}

func Del(key string) error {
	return Client.Del(ctx, key).Err()
}
