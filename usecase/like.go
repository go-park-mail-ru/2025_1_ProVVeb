package usecase

import (
	"fmt"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type ProfileSetLike struct {
	userRepo repository.UserRepository
}

func NewProfileSetLikeUseCase(userRepo repository.UserRepository) (*ProfileSetLike, error) {
	if userRepo == nil {
		return nil, fmt.Errorf("userRepo is nil")
	}
	return &ProfileSetLike{userRepo: userRepo}, nil
}

func (l *ProfileSetLike) SetLike(from int, to int, status int) (int, error) {
	return l.userRepo.SetLike(from, to, status)
}
