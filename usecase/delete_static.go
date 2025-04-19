package usecase

import (
	"fmt"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type DeleteStatic struct {
	userRepo   repository.UserRepository
	staticRepo repository.StaticRepository
}

func NewDeleteStaticUseCase(
	userRepo repository.UserRepository,
	staticRepo repository.StaticRepository,
) (*DeleteStatic, error) {
	if userRepo == nil || staticRepo == nil {
		return nil, fmt.Errorf("userRepo or staticRepo is nil")
	}
	return &DeleteStatic{userRepo: userRepo, staticRepo: staticRepo}, nil
}

func (su *DeleteStatic) DeleteImage(user_id int, filename string) error {
	err := su.staticRepo.DeleteImage(user_id, filename)

	if err != nil {
		return err
	}

	return su.userRepo.DeletePhoto(user_id, filename)
}
