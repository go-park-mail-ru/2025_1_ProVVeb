package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/sirupsen/logrus"
)

type CreateMessages struct {
	chatRepo repository.ChatRepository
	logger   *logger.LogrusLogger
}

func NewCreateMessagesUseCase(chatRepo repository.ChatRepository, logger *logger.LogrusLogger) (*CreateMessages, error) {
	return &CreateMessages{chatRepo: chatRepo, logger: logger}, nil
}

func (gp *CreateMessages) CreateMessages(chatID int, userID int, content string) (int, error) {
	gp.logger.Info("GetMessages", "chatID", chatID, "userID", userID, "content", content)
	messageID, err := gp.chatRepo.CreateMessage(chatID, userID, content, 1)
	if err != nil {
		gp.logger.Error("GetMessages", "chatID", chatID, "messageID", messageID, "error", err)
	} else {
		gp.logger.WithFields(&logrus.Fields{"chatID": chatID, "messageID": messageID})
	}
	return messageID, err
}
