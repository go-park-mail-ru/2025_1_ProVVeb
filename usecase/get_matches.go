package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type GetProfileMatches struct {
	userRepo repository.UserRepository
}

func NewGetProfileMatchesUseCase(userRepo repository.UserRepository) (*GetProfileMatches, error) {
	if userRepo == nil {
		return nil, model.ErrGetProfileMatchesUC
	}
	return &GetProfileMatches{userRepo: userRepo}, nil
}

func (gp *GetProfileMatches) GetMatches(forUserId int) ([]model.Profile, error) {
	return gp.userRepo.GetMatches(forUserId)
}
