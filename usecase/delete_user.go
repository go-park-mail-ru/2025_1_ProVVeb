package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/sirupsen/logrus"
)

type UserDelete struct {
	userRepo repository.UserRepository
	logger   *logger.LogrusLogger
}

func NewUserDeleteUseCase(userRepo repository.UserRepository, logger *logger.LogrusLogger) (*UserDelete, error) {
	if userRepo == nil || logger == nil {
		return nil, model.ErrUserDeleteUC
	}
	return &UserDelete{userRepo: userRepo, logger: logger}, nil
}

func (ud *UserDelete) DeleteUser(userId int) error {
	ud.logger.Info("DeleteUser", "userId", userId)
	err := ud.userRepo.DeleteUserById(userId)
	ud.logger.WithFields(&logrus.Fields{"userId": userId, "error": err})
	return err
}
