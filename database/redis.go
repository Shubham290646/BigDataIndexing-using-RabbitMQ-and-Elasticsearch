package database

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	redis "github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client redis.Client
}

func NewRedisRepository(address string) *RedisRepository {
	return &RedisRepository{
		client: *redis.NewClient(&redis.Options{
			Addr: address,
		}),
	}
}

func (r *RedisRepository) Ping(ctx *gin.Context) error {
	_, err := r.client.Ping(ctx).Result()
	return err
}

func (r *RedisRepository) Get(ctx *gin.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", errors.New("KEY_NOT_FOUND")
	}
	return val, err
}

func (r *RedisRepository) Set(ctx *gin.Context, key, value string) error {
	_, err := r.client.Set(ctx, key, value, 7*time.Hour).Result()
	return err
}

func (r *RedisRepository) Delete(ctx *gin.Context, key string) error {
	res, err := r.client.Del(ctx, key).Result()
	if res == 0 {
		return errors.New("KEY_NOT_FOUND")
	}
	return err
}

func (r *RedisRepository) Keys(ctx *gin.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}
