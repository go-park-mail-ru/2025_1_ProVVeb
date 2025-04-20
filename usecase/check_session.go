package usecase

import (
	"strconv"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type UserCheckSession struct {
	sessionRepo repository.SessionRepository
}

func NewUserCheckSessionUseCase(sessionRepo repository.SessionRepository) (*UserCheckSession, error) {
	if sessionRepo == nil {
		return nil, model.ErrUserCheckSessionUC
	}
	return &UserCheckSession{sessionRepo: sessionRepo}, nil
}

func (uc *UserCheckSession) CheckSession(sessionId string) (int, error) {
	userIdStr, err := uc.sessionRepo.GetSession(sessionId)
	if err != nil {
		return -1, err
	}

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return -1, model.ErrInvalidSessionId
	}
	return userId, nil
}
