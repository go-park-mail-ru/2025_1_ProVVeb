package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/sirupsen/logrus"
)

type ProfileUpdate struct {
	userRepo repository.UserRepository
	logger   *logger.LogrusLogger
}

func NewProfileUpdateUseCase(
	userRepo repository.UserRepository,
	logger *logger.LogrusLogger,
) (*ProfileUpdate, error) {
	if userRepo == nil || logger == nil {
		return nil, model.ErrProfileUpdateUC
	}
	return &ProfileUpdate{userRepo: userRepo, logger: logger}, nil
}

func (pu *ProfileUpdate) UpdateProfile(value model.Profile, targ model.Profile, profileId int) error {
	pu.logger.WithFields(&logrus.Fields{"profileId": profileId, "value profile": value}).Info("UpdateProfile")
	if value.FirstName != "" {
		targ.FirstName = value.FirstName
	}
	if value.LastName != "" {
		targ.LastName = value.LastName
	}
	if value.IsMale {
		targ.IsMale = !targ.IsMale
	}
	if value.Height != 0 {
		targ.Height = value.Height
	}
	if !value.Birthday.IsZero() {
		targ.Birthday = value.Birthday
	}
	if value.Description != "" {
		targ.Description = value.Description
	}
	if value.Location != "" {
		targ.Location = value.Location
	}

	if len(value.Interests) != 0 {
		targ.Interests = value.Interests
	}

	if len(value.Preferences) != 0 {
		targ.Preferences = value.Preferences
	}

	err := pu.userRepo.UpdateProfile(profileId, targ)
	pu.logger.WithFields(&logrus.Fields{"error": err}).Error("UpdateProfile")
	return err
}
