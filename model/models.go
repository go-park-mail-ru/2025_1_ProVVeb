package model

import (
	"errors"
	"time"
)

var MinPasswordLength = 8
var MaxPasswordLength = 64
var MinLoginLength = 7
var MaxLoginLength = 15

var SessionDuration = 3 * 24 * time.Hour
var SessionIdLength = 32

var PageSize = 5

// errors
var (
	ErrInvalidLogin          = errors.New("invalid login")
	ErrInvalidPassword       = errors.New("invalid password")
	ErrSessionNotFound       = errors.New("session not found")
	ErrInvalidSession        = errors.New("invalid session")
	ErrInvalidUserRepoConfig = errors.New("invalid user repository config")
	ErrGetSession            = errors.New("failed to get session")
	ErrStoreSession          = errors.New("failed to store session")
	ErrInvalidSessionId      = errors.New("invalid session id")
	ErrDeleteSession         = errors.New("failed to delete session")
	ErrProfileNotFound       = errors.New("profile not found")
	ErrDeleteUser            = errors.New("failed to delete user")
	ErrDeleteProfile         = errors.New("failed to delete profile")
)

type User struct {
	UserId   int    `yaml:"id" json:"id"`
	Login    string `yaml:"login" json:"login"`
	Password string `yaml:"password" json:"password"`
	Email    string `yaml:"email" json:"email"`
	Phone    string `yaml:"phone" json:"phone"`
	Status   int    `yaml:"status" json:"status"`
}

type Profile struct {
	ProfileId   int       `yaml:"profileId" json:"profileId"`
	FirstName   string    `yaml:"firstName" json:"firstName"`
	LastName    string    `yaml:"lastName" json:"lastName"`
	IsMale      bool      `yaml:"isMale" json:"isMale"`
	Height      int       `yaml:"height" json:"height"`
	Birthday    time.Time `yaml:"birthday" json:"birthday"`
	Avatar      string    `yaml:"avatar" json:"avatar"`
	Card        string    `yaml:"card" json:"card"`
	Description string    `yaml:"description" json:"description"`
	Location    string    `yaml:"location" json:"location"`
	Interests   []string  `yaml:"interests" json:"interests"`
	LikedBy     []int     `yaml:"likedBy" json:"likedBy"`
	Preferences []string  `yaml:"preferences" json:"preferences"`
}

type Session struct {
	SessionId string        `yaml:"sessionId" json:"sessionId"`
	UserId    int           `yaml:"userId" json:"userId"`
	Expires   time.Duration `yaml:"expires" json:"expires"`
}

type Cookie struct {
	Name     string    `yaml:"name" json:"name"`
	Value    string    `yaml:"value" json:"value"`
	Expires  time.Time `yaml:"expires" json:"expires"`
	HttpOnly bool      `yaml:"httpOnly" json:"httpOnly"`
	Secure   bool      `yaml:"secure" json:"secure"`
	Path     string    `yaml:"path" json:"path"`
}
