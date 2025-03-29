package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

func DBInitRedisConfig(redisAddr string, redisDB int) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       redisDB,
	})

	ctx := context.Background()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Ошибка подключения к Redis: %v", err)
	}

	return &RedisClient{
		client: client,
		ctx:    ctx,
	}
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

func (r *RedisClient) StoreSession(sessionID string, data string, ttl time.Duration) error {
	err := r.client.Set(r.ctx, sessionID, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("не удалось сохранить сессию в Redis: %v", err)
	}
	return nil
}

func (r *RedisClient) GetSession(sessionID string) (string, error) {
	data, err := r.client.Get(r.ctx, sessionID).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("сессия не найдена")
		}
		return "", fmt.Errorf("не удалось получить сессию из Redis: %v", err)
	}
	return data, nil
}

func (r *RedisClient) DeleteSession(sessionID string) error {
	err := r.client.Del(r.ctx, sessionID).Err()
	if err != nil {
		return fmt.Errorf("не удалось удалить сессию из Redis: %v", err)
	}
	return nil
}
