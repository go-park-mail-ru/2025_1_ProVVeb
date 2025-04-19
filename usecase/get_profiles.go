package usecase

import (
	"fmt"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type GetProfilesForUser struct {
	userRepo   repository.UserRepository
	staticRepo repository.StaticRepository
}

func NewGetProfilesForUserUseCase(
	userRepo repository.UserRepository,
	staticRepo repository.StaticRepository,
) (*GetProfilesForUser, error) {
	if userRepo == nil || staticRepo == nil {
		return nil, fmt.Errorf("userRepo or staticRepo undefined")
	}

	return &GetProfilesForUser{
		userRepo:   userRepo,
		staticRepo: staticRepo,
	}, nil
}

func (gp *GetProfilesForUser) GetProfiles(forUserId int) ([]model.Profile, error) {
	return gp.userRepo.GetProfilesByUserId(forUserId)
}
