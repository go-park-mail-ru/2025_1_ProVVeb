package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/sirupsen/logrus"
)

type StaticUpload struct {
	userRepo   repository.UserRepository
	staticRepo repository.StaticRepository
	logger     *logger.LogrusLogger
}

func NewStaticUploadUseCase(
	userRepo repository.UserRepository,
	staticRepo repository.StaticRepository,
	logger *logger.LogrusLogger,
) (*StaticUpload, error) {
	if userRepo == nil || staticRepo == nil || logger == nil {
		return nil, model.ErrStaticUploadUC
	}
	return &StaticUpload{userRepo: userRepo, staticRepo: staticRepo, logger: logger}, nil
}

func (su *StaticUpload) UploadUserPhoto(user_id int, file []byte, filename string, content_type string) error {
	su.logger.Info("UploadUserPhoto")
	filename_with_path := "/" + filename
	err := su.staticRepo.UploadImage(file, filename_with_path, content_type)

	su.logger.WithFields(&logrus.Fields{"user_id": user_id, "filename": filename, "content_type": content_type})
	if err != nil {
		su.logger.Error("UploadUserPhoto", err)
		return err
	}

	su.logger.Info("UploadUserPhoto: image uploaded")
	err = su.userRepo.StorePhoto(user_id, filename)

	su.logger.Info("error", err)
	return err
}
