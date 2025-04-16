package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type StaticUpload struct {
	userRepo   repository.UserRepository
	staticRepo repository.StaticRepository
}

func NewStaticUseCase(userRepo repository.UserRepository, staticRepo repository.StaticRepository) *StaticUpload {
	return &StaticUpload{userRepo: userRepo, staticRepo: staticRepo}
}

func (su *StaticUpload) UploadUserPhoto(user_id int, file []byte, filename string, content_type string) error {
	err := su.staticRepo.UploadImages(file, filename, content_type)

	if err != nil {
		return err
	}

	err = su.userRepo.StorePhoto(user_id, filename)
	if err != nil {
		return err
	}

	return nil
}
