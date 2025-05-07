package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/sirupsen/logrus"
)

type GetMessagesFromCache struct {
	chatRepo repository.ChatRepository
	logger   *logger.LogrusLogger
}

func NewGetMessagesFromCacheUseCase(chatRepo repository.ChatRepository, logger *logger.LogrusLogger) (*GetMessagesFromCache, error) {
	return &GetMessagesFromCache{chatRepo: chatRepo, logger: logger}, nil
}

func (gp *GetMessagesFromCache) GetMessages(chatID int, userID int) ([]model.Message, error) {
	gp.logger.Info("GetMessages", "chatID", chatID)
	messages, err := gp.chatRepo.GetMessagesFromCache(chatID, userID)
	if err != nil {
		gp.logger.Error("GetMessages", "chatID", chatID, "error", err)
	} else {
		gp.logger.WithFields(&logrus.Fields{"chatID": chatID, "messages": messages})
	}
	return messages, err
}
