package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	profilespb "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/sirupsen/logrus"
)

type GetUserPhoto struct {
	ProfilesService profilespb.ProfilesServiceClient
	logger          *logger.LogrusLogger
}

func NewGetUserPhotoUseCase(
	ProfilesService profilespb.ProfilesServiceClient,
	logger *logger.LogrusLogger,
) (*GetUserPhoto, error) {
	if ProfilesService == nil || logger == nil {
		return nil, model.ErrGetUserPhotoUC
	}
	return &GetUserPhoto{
		ProfilesService: ProfilesService,
		logger:          logger,
	}, nil
}

func (gp *GetUserPhoto) GetUserPhoto(user_id int) ([][]byte, []string, error) {
	gp.logger.Info("GetUserPhotoUseCase")
	req := &profilespb.GetProfileImagesRequest{
		UserId: int32(user_id),
	}

	res, err := gp.ProfilesService.GetProfileImages(context.Background(), req)
	if err != nil {
		gp.logger.WithFields(&logrus.Fields{
			"error": err,
		}).Error("GetUserPhotoUseCase")
		return nil, nil, err
	}

	var photos [][]byte
	var filenames []string
	for _, photo := range res.Files {
		photos = append(photos, photo)
	}
	for _, filename := range res.Urls {
		filenames = append(filenames, filename)
	}
	gp.logger.WithFields(&logrus.Fields{
		"len filenames": len(filenames),
	}).Info("GetUserPhotoUseCase")
	return photos, filenames, nil
}
