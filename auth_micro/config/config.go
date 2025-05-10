package auth_config

import (
	"errors"
	"time"
)

var SessionIdLength = 32
var SessionDuration = 3 * 24 * time.Hour

var (
	ErrSessionNotFound  = errors.New("session not found")
	ErrGetSession       = errors.New("failed to get session")
	ErrStoreSession     = errors.New("failed to store session")
	ErrInvalidSessionId = errors.New("invalid session id")
	ErrDeleteSession    = errors.New("failed to delete session")
)

const (
	AttemptsKeyPrefix     = "login_attempts:"
	TimeAttemptsKeyPrefix = "time:"
	MaxAttempts           = 5
	AttemptTTL            = 10 * time.Minute
)

type Session struct {
	SessionId string        `yaml:"sessionId" json:"sessionId"`
	UserId    int           `yaml:"userId" json:"userId"`
	Expires   time.Duration `yaml:"expires" json:"expires"`
}
