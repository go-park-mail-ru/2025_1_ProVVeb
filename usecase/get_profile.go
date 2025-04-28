package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/sirupsen/logrus"
)

type GetProfile struct {
	userRepo   repository.UserRepository
	staticRepo repository.StaticRepository
	logger     *logger.LogrusLogger
}

func NewGetProfileUseCase(
	userRepo repository.UserRepository,
	staticRepo repository.StaticRepository,
	logger *logger.LogrusLogger,
) (*GetProfile, error) {
	if userRepo == nil || staticRepo == nil || logger == nil {
		return &GetProfile{}, model.ErrGetProfileUC
	}
	return &GetProfile{userRepo: userRepo, staticRepo: staticRepo, logger: logger}, nil
}

func (gp *GetProfile) GetProfile(userId int) (model.Profile, error) {
	gp.logger.Info("GetProfile", "userId", userId)
	profile, err := gp.userRepo.GetProfileById(userId)
	if err != nil {
		gp.logger.Error("GetProfile", "userId", userId, "error", err)
	} else {
		gp.logger.WithFields(&logrus.Fields{"userId": userId, "profile": profile})
	}
	return profile, err
}

type GetUserPhoto struct {
	userRepo   repository.UserRepository
	staticRepo repository.StaticRepository
	logger     *logger.LogrusLogger
}

func NewGetUserPhotoUseCase(
	userRepo repository.UserRepository,
	staticRepo repository.StaticRepository,
	logger *logger.LogrusLogger,
) (*GetUserPhoto, error) {
	if userRepo == nil || staticRepo == nil || logger == nil {
		return nil, model.ErrGetUserPhotoUC
	}
	return &GetUserPhoto{
		userRepo:   userRepo,
		staticRepo: staticRepo,
		logger:     logger,
	}, nil
}

func (gp *GetUserPhoto) GetUserPhoto(user_id int) ([][]byte, []string, error) {
	gp.logger.Info("GetUserPhoto", "user_id", user_id)
	urls, err := gp.userRepo.GetPhotos(user_id)
	if err != nil {
		gp.logger.Error("GetUserPhoto", "user_id", user_id, "urls error", err)
		return nil, nil, err
	}

	gp.logger.Info("GetUserPhoto", "user_id", user_id, "urls", urls)
	files, err := gp.staticRepo.GetImages(urls)
	if err != nil {
		gp.logger.Error("GetUserPhoto", "user_id", user_id, "files error", err)
		return nil, nil, err
	}
	gp.logger.WithFields(&logrus.Fields{"user_id": user_id, "urlsCount": len(urls)}).Info("ok")
	return files, urls, err
}
