package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/sirupsen/logrus"
)

type GetProfilesForUser struct {
	userRepo   repository.UserRepository
	staticRepo repository.StaticRepository
	logger     *logger.LogrusLogger
}

func NewGetProfilesForUserUseCase(
	userRepo repository.UserRepository,
	staticRepo repository.StaticRepository,
	logger *logger.LogrusLogger,
) (*GetProfilesForUser, error) {
	if userRepo == nil || staticRepo == nil || logger == nil {
		return nil, model.ErrGetProfilesForUserUC
	}

	return &GetProfilesForUser{
		userRepo:   userRepo,
		staticRepo: staticRepo,
		logger:     logger,
	}, nil
}

func (gp *GetProfilesForUser) GetProfiles(forUserId int) ([]model.Profile, error) {
	gp.logger.Info("GetProfiles", "forUserId", forUserId)
	result, err := gp.userRepo.GetProfilesByUserId(forUserId)
	if err != nil {
		gp.logger.Error("GetProfiles", "forUserId", forUserId, "error", err)
	} else {
		gp.logger.WithFields(&logrus.Fields{"forUserId": forUserId, "profilesCount": len(result)}).Info("GetProfiles")
	}
	return result, err
}
