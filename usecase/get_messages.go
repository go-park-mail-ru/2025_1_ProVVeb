package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/sirupsen/logrus"
)

type GetMessages struct {
	chatRepo repository.ChatRepository
	logger   *logger.LogrusLogger
}

func NewGetMessagesUseCase(chatRepo repository.ChatRepository, logger *logger.LogrusLogger) (*GetMessages, error) {
	return &GetMessages{chatRepo: chatRepo, logger: logger}, nil
}

func (gp *GetMessages) GetMessages(chatID int) ([]model.Message, error) {
	gp.logger.Info("GetMessages", "chatID", chatID)
	messages, err := gp.chatRepo.GetMessages(chatID)
	if err != nil {
		gp.logger.Error("GetMessages", "chatID", chatID, "error", err)
	} else {
		gp.logger.WithFields(&logrus.Fields{"chatID": chatID, "messages": messages})
	}
	return messages, err
}
