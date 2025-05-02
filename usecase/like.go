package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/sirupsen/logrus"
)

type ProfileSetLike struct {
	userRepo repository.UserRepository
	logger   *logger.LogrusLogger
}

func NewProfileSetLikeUseCase(userRepo repository.UserRepository, logger *logger.LogrusLogger) (*ProfileSetLike, error) {
	if userRepo == nil || logger == nil {
		return nil, model.ErrProfileSetLikeUC
	}
	return &ProfileSetLike{userRepo: userRepo, logger: logger}, nil
}

func (l *ProfileSetLike) SetLike(from int, to int, status int) (int, error) {
	l.logger.WithFields(&logrus.Fields{"from": from, "to": to, "status": status}).Info("SetLike")
	result, err := l.userRepo.SetLike(from, to, status)
	if err != nil {
		l.logger.Error("SetLike", "error", err)
	} else {
		l.logger.Info("SetLike", "result", result)
	}
	return result, err
}
