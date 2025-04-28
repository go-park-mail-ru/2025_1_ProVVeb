package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/sirupsen/logrus"
)

type GetProfileMatches struct {
	userRepo repository.UserRepository
	logger   *logger.LogrusLogger
}

func NewGetProfileMatchesUseCase(userRepo repository.UserRepository, logger *logger.LogrusLogger) (*GetProfileMatches, error) {
	if userRepo == nil {
		return nil, model.ErrGetProfileMatchesUC
	}
	return &GetProfileMatches{userRepo: userRepo, logger: logger}, nil
}

func (gp *GetProfileMatches) GetMatches(forUserId int) ([]model.Profile, error) {
	gp.logger.Info("GetMatches", "forUserId", forUserId)
	result, err := gp.userRepo.GetMatches(forUserId)
	if err != nil {
		gp.logger.WithFields(&logrus.Fields{"forUserId": forUserId, "error": err}).Error("GetMatches", "error")
	} else {
		gp.logger.WithFields(&logrus.Fields{"forUserId": forUserId, "dataCount": len(result), "error": err})
	}

	return result, err
}
