package usecase

import (
	"strconv"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type UserLogOut struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	logger      *logger.LogrusLogger
}

func NewUserLogOutUseCase(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	logger *logger.LogrusLogger,
) (*UserLogOut, error) {
	if userRepo == nil || sessionRepo == nil || logger == nil {
		return nil, model.ErrUserLogOutUC
	}
	return &UserLogOut{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		logger:      logger,
	}, nil
}

func (ul *UserLogOut) Logout(sessionId string) error {
	ul.logger.Info("Logout", "sessionId", sessionId)
	userIdStr, err := ul.sessionRepo.GetSession(sessionId)
	if err != nil {
		ul.logger.Error("Logout", "sessionId", sessionId, "error", err)
		return err
	}

	ul.logger.Info("Logout", "userId", userIdStr)
	err = ul.sessionRepo.DeleteSession(sessionId)
	if err != nil {
		ul.logger.Error("Logout", "sessionId", sessionId, "error", err)
		return model.ErrDeleteSession
	}

	ul.logger.Info("Logout", "userId", userIdStr)
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		ul.logger.Error("Logout", "userId", userIdStr, "error", err)
		return model.ErrInvalidSessionId
	}

	ul.logger.Info("Logout", "userId", userId)
	err = ul.userRepo.DeleteSession(userId)
	if err != nil {
		ul.logger.Error("Logout", "userId", userId, "error", err)
		return model.ErrDeleteSession
	}
	ul.logger.Info("Logout", "ok", userId)

	return nil
}
