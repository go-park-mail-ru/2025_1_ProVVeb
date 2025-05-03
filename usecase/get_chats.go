package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"

	"github.com/sirupsen/logrus"
)

type GetChats struct {
	chatRepo repository.ChatRepository
	logger   *logger.LogrusLogger
}

func NewGetChatsUseCase(chatRepo repository.ChatRepository, logger *logger.LogrusLogger) (*GetChats, error) {

	return &GetChats{chatRepo: chatRepo, logger: logger}, nil
}

func (uc *GetChats) GetChats(userID int) ([]model.Chat, error) {
	uc.logger.Info("GetChats", "userId", userID)
	chats, err := uc.chatRepo.GetChats(userID)
	if err != nil {
		uc.logger.Error("GetProfile", "userId", userID, "error", err)
	} else {
		uc.logger.WithFields(&logrus.Fields{"userId": userID, "chats": chats})
	}
	return chats, err
}
