package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type ProfileSetLike struct {
	userRepo repository.UserRepository
}

func NewProfileLikeCase(userRepo repository.UserRepository) *ProfileSetLike {
	return &ProfileSetLike{userRepo: userRepo}
}

func (l *ProfileSetLike) SetLike(from int, to int, status int) (int, error) {
	return l.userRepo.SetLike(from, to, status)
}
