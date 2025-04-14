package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type ProfileGetMatches struct {
	userRepo repository.UserRepository
}

func NewProfileMatchCase(userRepo repository.UserRepository) *ProfileGetMatches {
	return &ProfileGetMatches{userRepo: userRepo}
}

func (gp *ProfileGetMatches) GetMatches(forUserId int) ([]model.Profile, error) {
	return gp.userRepo.GetMatches(forUserId)
}
