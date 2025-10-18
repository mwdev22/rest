package database

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	GetJSON(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	SetJSON(ctx context.Context, key string, value []byte, expiration time.Duration) error
	Keys(ctx context.Context, pattern string) ([]string, error)
	Del(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) bool
}

type RedisCache struct {
	Redis *redis.Client
}

func NewRedisCache(db int, addr string) Cache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}
	return &RedisCache{
		Redis: rdb,
	}
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return c.Redis.Get(ctx, key).Result()
}

func (c *RedisCache) GetJSON(ctx context.Context, key string) ([]byte, error) {
	value, err := c.Redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return []byte(value), nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.Redis.Set(ctx, key, value, expiration).Err()
}

func (c *RedisCache) SetJSON(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return c.Redis.Set(ctx, key, value, expiration).Err()
}

func (c *RedisCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.Redis.Keys(ctx, pattern).Result()
}

func (c *RedisCache) Del(ctx context.Context, key string) error {
	return c.Redis.Del(ctx, key).Err()
}

func (c *RedisCache) Exists(ctx context.Context, key string) bool {
	return c.Redis.Exists(ctx, key).Val() > 0
}
