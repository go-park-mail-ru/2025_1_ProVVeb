package repository

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-redis/redis/v8"
)

type SessionRepository interface {
	CreateSession(userId int) model.Session
	GetSession(sessionId string) (string, error)
	StoreSession(sessionId string, data string, ttl time.Duration) error
	DeleteSession(sessionId string) error
	CloseRepo() error

	CheckAttempts(userIP string) (string, error)
	IncreaseAttempts(userIP string) error
	DeleteAttempts(userIP string) error
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

func RandStringRunes(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (sr *SessionRepo) CreateSession(userId int) model.Session {
	session_id := RandStringRunes(model.SessionIdLength)
	expires := model.SessionDuration

	return model.Session{
		SessionId: session_id,
		UserId:    userId,
		Expires:   expires,
	}
}

func (sr *SessionRepo) DeleteSession(sessionId string) error {
	return sr.client.Del(sr.ctx, sessionId).Err()
}

func (sr *SessionRepo) GetSession(sessionId string) (string, error) {
	data, err := sr.client.Get(sr.ctx, sessionId).Result()
	if err != nil {
		if err == redis.Nil {
			return "", model.ErrSessionNotFound
		}
		return "", model.ErrGetSession
	}
	return data, nil
}

func (sr *SessionRepo) StoreSession(sessionId string, data string, ttl time.Duration) error {
	err := sr.client.Set(sr.ctx, sessionId, data, ttl).Err()
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

func (sr *SessionRepo) CheckAttempts(userIP string) (string, error) {
	tsKey := model.AttemptsKeyPrefix + userIP
	timeKey := model.TimeAttemptsKeyPrefix + userIP

	countStr, err := sr.client.Get(sr.ctx, tsKey).Result()
	if err == redis.Nil {
		if err := sr.client.Set(sr.ctx, tsKey, 0, model.AttemptTTL).Err(); err != nil {
			return "", err
		}
		return "", nil
	} else if err != nil {
		return "", err
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		return "", err
	}

	blockUntilStr, err := sr.client.Get(sr.ctx, timeKey).Result()
	fmt.Println(blockUntilStr, count, userIP)
	if err != nil && err != redis.Nil {
		return "", err
	}
	if blockUntilStr != "" {
		blockUntil, err := strconv.ParseInt(blockUntilStr, 10, 64)
		if err != nil {
			return "", err
		}
		if time.Now().Unix() < blockUntil {
			return blockUntilStr, errors.New("too many login attempts, try later")
		}
	}

	if count >= model.MaxAttempts {
		return "", errors.New("too many login attempts, try later")
	}

	return "", nil
}
func (sr *SessionRepo) IncreaseAttempts(userIP string) error {
	tsKey := model.AttemptsKeyPrefix + userIP
	timeKey := model.TimeAttemptsKeyPrefix + userIP

	count, err := sr.client.Incr(sr.ctx, tsKey).Result()
	if err != nil {
		return err
	}

	if count > model.MaxAttempts {
		additionalDelay := model.AttemptTTL * time.Duration(count-model.MaxAttempts)
		blockUntil := time.Now().Unix() + int64(additionalDelay.Seconds())
		return sr.client.Set(sr.ctx, timeKey, blockUntil, additionalDelay).Err()
	}

	return nil
}

func (sr *SessionRepo) DeleteAttempts(userIP string) error {
	tsKey := model.AttemptsKeyPrefix + userIP
	timeKey := model.TimeAttemptsKeyPrefix + userIP
	return sr.client.Del(sr.ctx, tsKey, timeKey).Err()
}
