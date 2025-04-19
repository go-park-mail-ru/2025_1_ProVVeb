package usecase

import (
	"fmt"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type UserDelete struct {
	userRepo repository.UserRepository
}

func NewUserDeleteUseCase(userRepo repository.UserRepository) (*UserDelete, error) {
	if userRepo == nil {
		return nil, fmt.Errorf("userRepo is nil")
	}
	return &UserDelete{userRepo: userRepo}, nil
}

func (ud *UserDelete) DeleteUser(userId int) error {
	return ud.userRepo.DeleteUserById(userId)
}
