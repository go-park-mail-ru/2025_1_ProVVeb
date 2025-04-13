package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type GetProfile struct {
	userRepo repository.UserRepository
}

func NewGetProfileUseCase(userRepo repository.UserRepository) *GetProfile {
	return &GetProfile{userRepo: userRepo}
}

func (gp *GetProfile) GetProfile(userId int) (model.Profile, error) {
	return gp.userRepo.GetProfileById(userId)
}
