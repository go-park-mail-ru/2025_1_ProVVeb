package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"

	"github.com/sirupsen/logrus"
)

type AddSubscription struct {
	subRepo repository.SubsriptionRepository
	logger  *logger.LogrusLogger
}

func NewAddSubscriptionUseCase(subRepo repository.SubsriptionRepository, logger *logger.LogrusLogger) (*AddSubscription, error) {

	return &AddSubscription{subRepo: subRepo, logger: logger}, nil
}

func (uc *AddSubscription) CreateSub(userID int, sub_type int, data string) error {
	uc.logger.Info("CreateSub", "userId", userID, "sub_type", sub_type)
	err := uc.subRepo.CreateSub(userID, sub_type, data)
	if err != nil {
		uc.logger.Error("CreateSub", "userId", userID, "error", err)
	} else {
		uc.logger.WithFields(&logrus.Fields{"AddNotification": userID})
	}
	return err
}
