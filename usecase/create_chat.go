package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"

	"github.com/sirupsen/logrus"
)

type CreateChat struct {
	chatRepo repository.ChatRepository
	logger   *logger.LogrusLogger
}

func NewCreateChatUseCase(chatRepo repository.ChatRepository, logger *logger.LogrusLogger) (*CreateChat, error) {

	return &CreateChat{chatRepo: chatRepo, logger: logger}, nil
}

func (uc *CreateChat) CreateChat(firstID int, secondID int) (int, error) {
	uc.logger.Info("CreateChat", "firstID", firstID, "secondID", secondID)

	chatID, err := uc.chatRepo.CreateChat(firstID, secondID)
	if err != nil {
		uc.logger.Error("CreateChat", "firstID", firstID, "secondID", secondID, "error", err)
	} else {
		uc.logger.WithFields(&logrus.Fields{"firstID": firstID, "secondID": secondID, "chatID": chatID})
	}
	return chatID, err
}
