package auth

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	auth_config "github.com/go-park-mail-ru/2025_1_ProVVeb/auth_micro/config"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5"
)

type SessionRepository interface {
	CreateSession(userId int) auth_config.Session
	DeleteSession(sessionId string) error
	GetSession(sessionId string) (string, error)
	StoreSession(sessionId string, data string, ttl time.Duration) error
	DeleteAllSessions() error
	CloseRepo() error
	CheckAttempts(userIP string) (string, error)
	IncreaseAttempts(userIP string) error
	DeleteAttempts(userIP string) error
}

func RandStringRunes(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (sr *SessionRepo) CreateSession(userId int) auth_config.Session {
	session_id := RandStringRunes(auth_config.SessionIdLength)
	expires := auth_config.SessionDuration

	return auth_config.Session{
		SessionId: session_id,
		UserId:    userId,
		Expires:   expires,
	}
}

const (
	FindSessionQuery = `
SELECT id FROM sessions WHERE user_id = $1;
`
	DeleteSessionQuery = `
DELETE FROM sessions WHERE user_id = $1;
`
)

func (sr *SessionRepo) DeleteSession(sessionId string) error {
	var profileId int
	userId, err := strconv.Atoi(sessionId)
	if err != nil {
		return auth_config.ErrInvalidSessionId
	}
	err = sr.DB.QueryRowContext(context.Background(), FindSessionQuery, userId).Scan(&profileId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return auth_config.ErrSessionNotFound
		}
		return auth_config.ErrDeleteSession
	}
	_, err = sr.DB.ExecContext(context.Background(), DeleteSessionQuery, userId)
	if err != nil {
		return auth_config.ErrDeleteSession
	}

	return sr.Client.Del(sr.Ctx, sessionId).Err()
}

func (sr *SessionRepo) GetSession(sessionId string) (string, error) {
	data, err := sr.Client.Get(sr.Ctx, sessionId).Result()
	if err != nil {
		if err == redis.Nil {
			return "", auth_config.ErrSessionNotFound
		}
		return "", auth_config.ErrGetSession
	}
	return data, nil
}

const StoreSessionQuery = `
INSERT INTO sessions (user_id, token, created_at, expires_at)
VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + INTERVAL '72 hours')
RETURNING id;
`

func (sr *SessionRepo) StoreSession(userID int, session_id string, token string, ttl time.Duration) error {
	var sessionId int

	err := sr.DB.QueryRowContext(
		context.Background(),
		StoreSessionQuery,
		userID,
		token,
	).Scan(&sessionId)

	if err != nil {
		return err
	}

	userIDStr := strconv.Itoa(userID)
	err = sr.Client.Set(sr.Ctx, session_id, userIDStr, ttl).Err()
	if err != nil {
		return auth_config.ErrStoreSession
	}

	return nil
}

func (sr *SessionRepo) DeleteAllSessions() error {
	return sr.Client.FlushAll(sr.Ctx).Err()
}

func (sr *SessionRepo) CloseRepo() error {
	var err error
	if sr.DB != nil {
		err = sr.DB.Close()
		if err != nil {
			fmt.Printf("failed while closing connection: %v\n", err)
		}
	}
	return sr.Client.Close()
}

func (sr *SessionRepo) CheckAttempts(userIP string) (string, error) {
	tsKey := auth_config.AttemptsKeyPrefix + userIP
	timeKey := auth_config.TimeAttemptsKeyPrefix + userIP

	countStr, err := sr.Client.Get(sr.Ctx, tsKey).Result()
	if err == redis.Nil {
		if err := sr.Client.Set(sr.Ctx, tsKey, 0, auth_config.AttemptTTL).Err(); err != nil {
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

	if count >= auth_config.MaxAttempts {
		return "", errors.New("too many login attempts, try later")
	}

	return "", nil
}
func (sr *SessionRepo) IncreaseAttempts(userIP string) error {
	tsKey := auth_config.AttemptsKeyPrefix + userIP
	timeKey := auth_config.TimeAttemptsKeyPrefix + userIP

	count, err := sr.Client.Incr(sr.Ctx, tsKey).Result()
	if err != nil {
		return err
	}

	if count >= auth_config.MaxAttempts {
		additionalDelay := auth_config.AttemptTTL * time.Duration(count-auth_config.MaxAttempts)
		blockUntil := time.Now().Unix() + int64(additionalDelay.Seconds())
		return sr.Client.Set(sr.Ctx, timeKey, blockUntil, additionalDelay).Err()
	}

	return nil
}

func (sr *SessionRepo) DeleteAttempts(userIP string) error {
	tsKey := auth_config.AttemptsKeyPrefix + userIP
	timeKey := auth_config.TimeAttemptsKeyPrefix + userIP
	return sr.Client.Del(sr.Ctx, tsKey, timeKey).Err()
}
