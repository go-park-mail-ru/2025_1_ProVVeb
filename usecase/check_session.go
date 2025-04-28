package usecase

import (
	"strconv"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/sirupsen/logrus"
)

type UserCheckSession struct {
	sessionRepo repository.SessionRepository
	logger      *logger.LogrusLogger
}

func NewUserCheckSessionUseCase(sessionRepo repository.SessionRepository, logger *logger.LogrusLogger) (*UserCheckSession, error) {
	if sessionRepo == nil || logger == nil {
		return nil, model.ErrUserCheckSessionUC
	}
	return &UserCheckSession{sessionRepo: sessionRepo, logger: logger}, nil
}

func (uc *UserCheckSession) CheckSession(sessionId string) (int, error) {
	uc.logger.Info("Checking session")
	userIdStr, err := uc.sessionRepo.GetSession(sessionId)
	if err != nil {
		uc.logger.Error("Session not found")
		return -1, err
	}
	uc.logger.Info("Session found")

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		uc.logger.Error("Invalid session id")
		return -1, model.ErrInvalidSessionId
	}
	uc.logger.WithFields(&logrus.Fields{"userId": userId}).Info("Session checked")
	return userId, nil
}
