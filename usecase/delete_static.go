package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/sirupsen/logrus"
)

type DeleteStatic struct {
	userRepo   repository.UserRepository
	staticRepo repository.StaticRepository
	logger     *logger.LogrusLogger
}

func NewDeleteStaticUseCase(
	userRepo repository.UserRepository,
	staticRepo repository.StaticRepository,
	logger *logger.LogrusLogger,
) (*DeleteStatic, error) {
	if userRepo == nil || staticRepo == nil || logger == nil {
		return nil, model.ErrDeleteStaticUC
	}
	return &DeleteStatic{userRepo: userRepo, staticRepo: staticRepo, logger: logger}, nil
}

func (su *DeleteStatic) DeleteImage(user_id int, filename string) error {
	su.logger.Info("DeleteImage", &logrus.Fields{"user_id": user_id, "filename": filename})
	err := su.staticRepo.DeleteImage(user_id, filename)

	if err != nil {
		su.logger.Error("DeleteImage", &logrus.Fields{"error": err})
		return err
	}

	su.logger.Info("DeleteImage: static image deleted")

	err = su.userRepo.DeletePhoto(user_id, filename)
	if err != nil {
		su.logger.Error("DeleteImage", &logrus.Fields{"error": err})
		return err
	}

	su.logger.Info("DeleteImage: user image deleted")
	return err
}
