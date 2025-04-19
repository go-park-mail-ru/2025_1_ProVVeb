package usecase

import (
	"fmt"
	"strconv"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type UserLogOut struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
}

func NewUserLogOutUseCase(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
) (*UserLogOut, error) {
	if userRepo == nil || sessionRepo == nil {
		return nil, fmt.Errorf("userRepo or sessionRepo is nil")
	}
	return &UserLogOut{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}, nil
}

func (ul *UserLogOut) Logout(sessionId string) error {
	userIdStr, err := ul.sessionRepo.GetSession(sessionId)
	if err != nil {
		return err
	}

	err = ul.sessionRepo.DeleteSession(sessionId)
	if err != nil {
		return model.ErrDeleteSession
	}

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return model.ErrInvalidSessionId
	}

	err = ul.userRepo.DeleteSession(userId)
	if err != nil {
		return model.ErrDeleteSession
	}

	return nil
}
