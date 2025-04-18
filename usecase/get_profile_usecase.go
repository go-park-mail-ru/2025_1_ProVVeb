package usecase

import (
	"fmt"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type GetProfile struct {
	userRepo   repository.UserRepository
	staticRepo repository.StaticRepository
}

func NewGetProfileUseCase(userRepo repository.UserRepository, staticRepo repository.StaticRepository) *GetProfile {
	return &GetProfile{userRepo: userRepo, staticRepo: staticRepo}
}

func (gp *GetProfile) GetProfile(userId int) (model.Profile, error) {
	profile, err := gp.userRepo.GetProfileById(userId)
	return profile, err
}

type GetUserPhoto struct {
	userRepo   repository.UserRepository
	staticRepo repository.StaticRepository
}

func NewGetUserPhotoUseCase(userRepo repository.UserRepository, staticRepo repository.StaticRepository) *GetUserPhoto {
	return &GetUserPhoto{userRepo: userRepo, staticRepo: staticRepo}
}

func (gp *GetUserPhoto) GetUserPhoto(user_id int) ([][]byte, []string, error) {
	urls, err := gp.userRepo.GetPhotos(user_id)
	fmt.Println(urls)
	if err != nil {
		fmt.Println("urls", err)
		return nil, nil, err
	}

	files, err := gp.staticRepo.GetImages(urls)
	if err != nil {
		fmt.Println("files", err)
		return nil, nil, err
	}
	return files, urls, err
}
