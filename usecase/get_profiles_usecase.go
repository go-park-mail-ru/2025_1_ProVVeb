package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type GetProfilesForUser struct {
	userRepo repository.UserRepository
}

func NewGetProfilesForUserUseCase(userRepo repository.UserRepository) *GetProfilesForUser {
	return &GetProfilesForUser{userRepo: userRepo}
}

func (gp *GetProfilesForUser) GetProfiles(forUserId int) ([]model.Profile, error) {
	return gp.userRepo.GetProfilesByUserId(forUserId)
}
