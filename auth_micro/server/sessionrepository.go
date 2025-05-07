package auth

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-redis/redis/v8"
)

type SessionRepository interface {
	CreateSession(userId int) model.Session
	DeleteSession(sessionId string) error
	GetSession(sessionId string) (string, error)
	StoreSession(sessionId string, data string, ttl time.Duration) error
	DeleteAllSessions() error
	CloseRepo() error
	CheckAttempts(userIP string) (string, error)
	IncreaseAttempts(userIP string) error
	DeleteAttempts(userIP string) error
}

type SessionRepo struct {
	Client *redis.Client
	Ctx    context.Context
}

func NewSessionRepo() (*SessionRepo, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return &SessionRepo{}, err
	}

	return &SessionRepo{
		Client: client,
		Ctx:    ctx,
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
	return sr.Client.Del(sr.Ctx, sessionId).Err()
}

func (sr *SessionRepo) GetSession(sessionId string) (string, error) {
	data, err := sr.Client.Get(sr.Ctx, sessionId).Result()
	if err != nil {
		if err == redis.Nil {
			return "", model.ErrSessionNotFound
		}
		return "", model.ErrGetSession
	}
	return data, nil
}

func (sr *SessionRepo) StoreSession(sessionId string, data string, ttl time.Duration) error {
	err := sr.Client.Set(sr.Ctx, sessionId, data, ttl).Err()
	if err != nil {
		return model.ErrStoreSession
	}
	return nil
}

func (sr *SessionRepo) DeleteAllSessions() error {
	return sr.Client.FlushAll(sr.Ctx).Err()
}

func (sr *SessionRepo) CloseRepo() error {
	return sr.Client.Close()
}

func (sr *SessionRepo) CheckAttempts(userIP string) (string, error) {
	tsKey := model.AttemptsKeyPrefix + userIP
	timeKey := model.TimeAttemptsKeyPrefix + userIP

	countStr, err := sr.Client.Get(sr.Ctx, tsKey).Result()
	if err == redis.Nil {
		if err := sr.Client.Set(sr.Ctx, tsKey, 0, model.AttemptTTL).Err(); err != nil {
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

	blockUntilStr, err := sr.Client.Get(sr.Ctx, timeKey).Result()
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

	count, err := sr.Client.Incr(sr.Ctx, tsKey).Result()
	if err != nil {
		return err
	}

	if count >= model.MaxAttempts {
		additionalDelay := model.AttemptTTL * time.Duration(count-model.MaxAttempts)
		blockUntil := time.Now().Unix() + int64(additionalDelay.Seconds())
		return sr.Client.Set(sr.Ctx, timeKey, blockUntil, additionalDelay).Err()
	}

	return nil
}

func (sr *SessionRepo) DeleteAttempts(userIP string) error {
	tsKey := model.AttemptsKeyPrefix + userIP
	timeKey := model.TimeAttemptsKeyPrefix + userIP
	return sr.Client.Del(sr.Ctx, tsKey, timeKey).Err()
}
