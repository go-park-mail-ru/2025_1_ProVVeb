package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/sirupsen/logrus"
)

type UpdateNotificationStatus struct {
	notifRepo repository.NotificationsRepository
	logger    *logger.LogrusLogger
}

func NewUpdateNotificationStatusUseCase(notifRepo repository.NotificationsRepository, logger *logger.LogrusLogger) (*UpdateNotificationStatus, error) {
	return &UpdateNotificationStatus{notifRepo: notifRepo, logger: logger}, nil
}

func (gp *UpdateNotificationStatus) UpdateNotificatons(userID int, nofit_type string) error {
	gp.logger.Info("UpdateNotificatons", "userID", userID)
	err := gp.notifRepo.MarkNotifications(userID, nofit_type)
	if err != nil {
		gp.logger.Error("UpdateNotificatons", "userID", userID, "error", err)
	} else {
		gp.logger.WithFields(&logrus.Fields{"userID": userID})
	}
	return err
}
