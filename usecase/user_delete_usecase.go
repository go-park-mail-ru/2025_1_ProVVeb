package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type UserDelete struct {
	userRepo repository.UserRepository
}

func NewUserDeleteUseCase(userRepo repository.UserRepository) *UserDelete {
	return &UserDelete{userRepo: userRepo}
}

func (ud *UserDelete) DeleteUser(userId int) error {
	return ud.userRepo.DeleteUserById(userId)
}
