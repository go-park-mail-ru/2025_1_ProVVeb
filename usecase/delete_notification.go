package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"

	"github.com/sirupsen/logrus"
)

type DeleteNotification struct {
	notifRepo repository.NotificationsRepository
	logger    *logger.LogrusLogger
}

func NewDeleteNotificationUseCase(notifRepo repository.NotificationsRepository, logger *logger.LogrusLogger) (*DeleteNotification, error) {

	return &DeleteNotification{notifRepo: notifRepo, logger: logger}, nil
}

func (uc *DeleteNotification) DeleteNotifications(notification_id int, userID int) error {
	uc.logger.Info("DeleteChat", "notification_id", notification_id, "userID", userID)

	err := uc.notifRepo.DeleteNotifications(notification_id, userID)
	if err != nil {
		uc.logger.Error("DeleteChat", "notification_id", notification_id, "userID", userID, "error", err)
	} else {
		uc.logger.WithFields(&logrus.Fields{"notification_id": notification_id, "userID": userID})
	}
	return err
}
