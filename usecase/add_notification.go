package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"

	"github.com/sirupsen/logrus"
)

type AddNotification struct {
	notifRepo repository.NotificationsRepository
	logger    *logger.LogrusLogger
}

func NewAddNotificationUseCase(notifRepo repository.NotificationsRepository, logger *logger.LogrusLogger) (*AddNotification, error) {

	return &AddNotification{notifRepo: notifRepo, logger: logger}, nil
}

func (uc *AddNotification) AddNotification(userID int, notif model.NotificationSend) error {
	uc.logger.Info("AddNotification", "userId", userID, "notif", notif)
	err := uc.notifRepo.AddNotification(userID, notif)
	if err != nil {
		uc.logger.Error("AddNotification", "userId", userID, "error", err)
	} else {
		uc.logger.WithFields(&logrus.Fields{"AddNotification": userID})
	}
	return err
}
