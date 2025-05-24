package model

import (
	"errors"
)

type User struct {
	UserId   int    `yaml:"id" json:"id"`
	Login    string `yaml:"login" json:"login"`
	Password string `yaml:"password" json:"password"`
	Email    string `yaml:"email" json:"email"`
	Phone    string `yaml:"phone" json:"phone"`
	Status   int    `yaml:"status" json:"status"`
}

var MinPasswordLength = 8
var MaxPasswordLength = 64
var MinLoginLength = 7
var MaxLoginLength = 25

var PageSize = 10
var MaxFileSize int64 = 10 << 20

const Megabyte int = 1 << 23
const MaxQuerySizeStr int = 5
const MaxQuerySizePhoto int = 15 * 6

// regexps
var (
	ReStartsWithLetter             = `^[a-zA-Z]`
	ReContainsLettersDigitsSymbols = `^[a-zA-Z0-9._-]+$`
)

var (
	ErrInvalidUserRepoConfig = errors.New("invalid user repository config")
	ErrDeleteUser            = errors.New("failed to delete user")
	ErrSessionNotFound       = errors.New("session not found")
	ErrDeleteSession         = errors.New("failed to delete session")
	ErrUserGetParamsUC       = errors.New("failed to get user params")
	ErrInvalidLogin          = errors.New("invalid login")
	ErrInvalidLoginSize      = errors.New("invalid login size")
	ErrInvalidPassword       = errors.New("invalid password")
	ErrInvalidPasswordSize   = errors.New("invalid password size")
)
