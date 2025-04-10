package storage

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage(addr, password string) *RedisStorage {
	return &RedisStorage{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       0,
		}),
	}
}

func (r *RedisStorage) Increment(key string, window time.Duration) (int64, error) {
	ctx := context.Background()

	// Script Lua para atomicidade (INCR + SET se não existir)
	script := `
	local current = redis.call("INCR", KEYS[1])
	if current == 1 then
		redis.call("EXPIRE", KEYS[1], ARGV[1])
	end
	return current
	`

	result, err := r.client.Eval(ctx, script, []string{key}, window.Seconds()).Int64()
	if err != nil {
		return 0, err
	}

	return result, nil
}

func (r *RedisStorage) IsBlocked(key string) (bool, error) {
	ctx := context.Background()
	exists, err := r.client.Exists(ctx, "blocked:"+key).Result()
	return exists == 1, err
}

func (r *RedisStorage) Block(key string, duration time.Duration) error {
	ctx := context.Background()
	return r.client.Set(ctx, "blocked:"+key, "1", duration).Err()
}
