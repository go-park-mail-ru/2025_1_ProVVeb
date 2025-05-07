package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/sirupsen/logrus"
)

type DeleteMessage struct {
	chatRepo repository.ChatRepository
	logger   *logger.LogrusLogger
}

func NewDeleteMessageUseCase(chatRepo repository.ChatRepository, logger *logger.LogrusLogger) (*DeleteMessage, error) {
	return &DeleteMessage{chatRepo: chatRepo, logger: logger}, nil
}

func (gp *DeleteMessage) DeleteMessage(messageID int, chatID int) error {
	gp.logger.Info("DeleteMessage", "chatID", chatID, "messageID", messageID)
	err := gp.chatRepo.DeleteMessage(messageID, chatID)
	if err != nil {
		gp.logger.Error("DeleteMessage", "chatID", chatID, "messageID", messageID, "error", err)
	} else {
		gp.logger.WithFields(&logrus.Fields{"chatID": chatID, "messageID": messageID})
	}
	return err
}
