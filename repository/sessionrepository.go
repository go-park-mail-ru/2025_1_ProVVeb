package repository

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-redis/redis/v8"
)

type SessionRepository interface {
	GetSession(sessionID string) (string, error)
	StoreSession(sessionID string, data string, ttl time.Duration) error
	CloseRepo() error
	DeleteSession(sessionID string) error
}

type SessionRepo struct {
	client *redis.Client
	ctx    context.Context
}

func NewSessionRepo() (*SessionRepo, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return &SessionRepo{}, err
	}

	return &SessionRepo{
		client: client,
		ctx:    ctx,
	}, nil
}

func (sr *SessionRepo) DeleteSession(sessionID string) error {
	return sr.client.Del(sr.ctx, sessionID).Err()
}

func (sr *SessionRepo) GetSession(sessionID string) (string, error) {
	data, err := sr.client.Get(sr.ctx, sessionID).Result()
	if err != nil {
		if err == redis.Nil {
			return "", model.ErrSessionNotFound
		}
		return "", model.ErrGetSession
	}
	return data, nil
}

func (sr *SessionRepo) StoreSession(sessionID string, data string, ttl time.Duration) error {
	err := sr.client.Set(sr.ctx, sessionID, data, ttl).Err()
	if err != nil {
		return model.ErrStoreSession
	}
	return nil
}

func (sr *SessionRepo) DeleteAllSessions() error {
	return sr.client.FlushAll(sr.ctx).Err()
}

func (sr *SessionRepo) CloseRepo() error {
	return sr.client.Close()
}
