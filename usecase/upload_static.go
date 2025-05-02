package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	profilespb "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/sirupsen/logrus"
)

type StaticUpload struct {
	ProfilesService profilespb.ProfilesServiceClient
	logger          *logger.LogrusLogger
}

func NewStaticUploadUseCase(
	ProfilesService profilespb.ProfilesServiceClient,
	logger *logger.LogrusLogger,
) (*StaticUpload, error) {
	if ProfilesService == nil || logger == nil {
		return nil, model.ErrStaticUploadUC
	}
	return &StaticUpload{ProfilesService: ProfilesService, logger: logger}, nil
}

func (su *StaticUpload) UploadUserPhoto(user_id int, file []byte, filename string, content_type string) error {
	su.logger.Info("StaticUploadUseCase")
	req := &profilespb.UploadProfileImageRequest{
		UserId:      int32(user_id),
		File:        file,
		Filename:    filename,
		ContentType: content_type,
	}
	_, err := su.ProfilesService.UploadProfileImage(context.Background(), req)
	su.logger.WithFields(&logrus.Fields{
		"upload profile image err": err,
	})
	return err
}
