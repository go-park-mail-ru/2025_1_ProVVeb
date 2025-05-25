package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"

	"github.com/sirupsen/logrus"
)

type GetNotifications struct {
	notifRepo repository.NotificationsRepository
	logger    *logger.LogrusLogger
}

func NewGetNotificationsUseCase(notifRepo repository.NotificationsRepository, logger *logger.LogrusLogger) (*GetNotifications, error) {

	return &GetNotifications{notifRepo: notifRepo, logger: logger}, nil
}

func (uc *GetNotifications) GetNotifications(userID int) ([]model.NotificationSend, error) {
	uc.logger.Info("GetNotifications", "userId", userID)
	notif, err := uc.notifRepo.GetNotifications(userID)
	if err != nil {
		uc.logger.Error("GetNotifications", "userId", userID, "error", err)
	} else {
		uc.logger.WithFields(&logrus.Fields{"userId": userID, "notif": notif})
	}
	return notif, err
}

type GetCurrentNotifications struct {
	notifRepo repository.NotificationsRepository
	logger    *logger.LogrusLogger
}

func NewGetCurrentNotificationsUseCase(notifRepo repository.NotificationsRepository, logger *logger.LogrusLogger) (*GetCurrentNotifications, error) {

	return &GetCurrentNotifications{notifRepo: notifRepo, logger: logger}, nil
}

func (uc *GetCurrentNotifications) GetCurrentNotifications(userID int) ([]model.NotificationSend, error) {
	uc.logger.Info("GetNotifications", "userId", userID)
	notif, err := uc.notifRepo.GetCurrentNotifications(userID)
	if err != nil {
		uc.logger.Error("GetNotifications", "userId", userID, "error", err)
	} else {
		uc.logger.WithFields(&logrus.Fields{"userId": userID, "notif": notif})
	}
	return notif, err
}
