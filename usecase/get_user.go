package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type UserGetParams struct {
	userRepo repository.UserRepository
}

func NewUserGetParamsUseCase(userRepo repository.UserRepository) (*UserGetParams, error) {
	return &UserGetParams{userRepo: userRepo}, nil
}

func (up *UserGetParams) GetUserParams(userId int) (model.User, error) {
	user, err := up.userRepo.GetUserParams(userId)
	return user, err
}
