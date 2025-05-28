package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"

	"github.com/sirupsen/logrus"
)

type UpdateBorder struct {
	subRepo repository.SubsriptionRepository
	logger  *logger.LogrusLogger
}

func NewUpdateBorderUseCase(subRepo repository.SubsriptionRepository, logger *logger.LogrusLogger) (*UpdateBorder, error) {

	return &UpdateBorder{subRepo: subRepo, logger: logger}, nil
}

func (uc *UpdateBorder) UpdateBorder(userID int, new_border int) error {
	uc.logger.Info("UpdateBorder", "userId", userID, "new_border", new_border)
	err := uc.subRepo.UpdateBorder(userID, new_border)
	if err != nil {
		uc.logger.Error("UpdateBorder", "userId", userID, "error", err)
	} else {
		uc.logger.WithFields(&logrus.Fields{"UpdateBorder": userID})
	}
	return err
}
