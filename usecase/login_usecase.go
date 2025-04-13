package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type LogInInput struct {
	Login    string
	Password string
}

type UserLogIn struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	hasher      repository.PasswordHasher
	validator   repository.UserParamsValidator
}

func NewUserLogInUseCase(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	hasher repository.PasswordHasher,
	validator repository.UserParamsValidator,
) *UserLogIn {
	return &UserLogIn{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		hasher:      hasher,
		validator:   validator,
	}
}
