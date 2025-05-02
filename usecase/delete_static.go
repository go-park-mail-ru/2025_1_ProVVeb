package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	profilespb "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/sirupsen/logrus"
)

type DeleteStatic struct {
	ProfilesService profilespb.ProfilesServiceClient
	logger         *logger.LogrusLogger
}

func NewDeleteStaticUseCase(
	ProfilesService profilespb.ProfilesServiceClient,
	logger *logger.LogrusLogger,
) (*DeleteStatic, error) {
	if ProfilesService == nil || logger == nil {
		return nil, model.ErrDeleteStaticUC
	}
	return &DeleteStatic{ProfilesService: ProfilesService, logger: logger}, nil
}

func (ds *DeleteStatic) DeleteImage(user_id int, filename string) error {
	ds.logger.WithFields(&logrus.Fields{
		"user_id":  user_id,
		"filename": filename,
	})
	req := &profilespb.DeleteImageRequest{
		UserId:   int32(user_id),
		Filename: filename,
	}

	_, err := ds.ProfilesService.DeleteImage(context.Background(), req)
	ds.logger.WithFields(&logrus.Fields{
		"user_id":  user_id,
		"filename": filename,
		"error":    err,
	})
	return err
}
