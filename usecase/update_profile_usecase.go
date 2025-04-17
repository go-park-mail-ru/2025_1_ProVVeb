package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type ProfileUpdate struct {
	userRepo repository.UserRepository
}

func NewProfileUpdateUseCase(userRepo repository.UserRepository) *ProfileUpdate {
	return &ProfileUpdate{userRepo: userRepo}
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

	return pu.userRepo.UpdateProfile(profileId, targ)
}
