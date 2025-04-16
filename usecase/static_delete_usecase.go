package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type StaticDelete struct {
	userRepo   repository.UserRepository
	staticRepo repository.StaticRepository
}

func NewStaticDeleteCase(userRepo repository.UserRepository, staticRepo repository.StaticRepository) *StaticDelete {
	return &StaticDelete{userRepo: userRepo, staticRepo: staticRepo}
}

func (su *StaticDelete) DeleteImage(user_id int, filename string) error {
	err := su.staticRepo.DeleteImage(user_id, filename)

	if err != nil {
		return err
	}

	return su.userRepo.DeletePhoto(user_id, filename)
}
