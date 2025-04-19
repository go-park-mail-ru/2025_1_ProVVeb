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

func NewGetProfileUseCase(
	userRepo repository.UserRepository,
	staticRepo repository.StaticRepository,
) (*GetProfile, error) {
	if userRepo == nil || staticRepo == nil {
		return &GetProfile{}, fmt.Errorf("userRepo or staticRepo undefined")
	}
	return &GetProfile{userRepo: userRepo, staticRepo: staticRepo}, nil
}

func (gp *GetProfile) GetProfile(userId int) (model.Profile, error) {
	profile, err := gp.userRepo.GetProfileById(userId)
	return profile, err
}

type GetUserPhoto struct {
	userRepo   repository.UserRepository
	staticRepo repository.StaticRepository
}

func NewGetUserPhotoUseCase(
	userRepo repository.UserRepository,
	staticRepo repository.StaticRepository,
) (*GetUserPhoto, error) {
	if userRepo == nil || staticRepo == nil {
		return nil, fmt.Errorf("userRepo or staticRepo undefined")
	}
	return &GetUserPhoto{
		userRepo:   userRepo,
		staticRepo: staticRepo,
	}, nil
}

func (gp *GetUserPhoto) GetUserPhoto(user_id int) ([][]byte, []string, error) {
	urls, err := gp.userRepo.GetPhotos(user_id)
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
