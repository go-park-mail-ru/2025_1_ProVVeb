package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type ProfileUpdate struct {
	userRepo repository.UserRepository
}

func NewProfileUpdateUseCase(userRepo repository.UserRepository) (*ProfileUpdate, error) {
	if userRepo == nil {
		return nil, model.ErrProfileUpdateUC
	}
	return &ProfileUpdate{userRepo: userRepo}, nil
}

func (pu *ProfileUpdate) UpdateProfile(value model.Profile, targ model.Profile, profileId int) error {
	if value.FirstName != "" {
		targ.FirstName = value.FirstName
	}
	if value.LastName != "" {
		targ.LastName = value.LastName
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

	return pu.userRepo.UpdateProfile(profileId, targ)
}
