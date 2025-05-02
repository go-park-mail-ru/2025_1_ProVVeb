package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/sirupsen/logrus"
)

type UpdateMessageStatus struct {
	chatRepo repository.ChatRepository
	logger   *logger.LogrusLogger
}

func NewUpdateMessageStatusUseCase(chatRepo repository.ChatRepository, logger *logger.LogrusLogger) (*UpdateMessageStatus, error) {
	return &UpdateMessageStatus{chatRepo: chatRepo, logger: logger}, nil
}

func (gp *UpdateMessageStatus) UpdateMessageStatus(chatID int, userID int) error {
	gp.logger.Info("GetMessages", "chatID", chatID)
	err := gp.chatRepo.UpdateMessageStatus(chatID, userID)
	if err != nil {
		gp.logger.Error("GetMessages", "chatID", chatID, "error", err)
	} else {
		gp.logger.WithFields(&logrus.Fields{"chatID": chatID})
	}
	return err
}
