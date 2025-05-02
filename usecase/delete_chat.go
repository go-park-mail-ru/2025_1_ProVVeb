package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"

	"github.com/sirupsen/logrus"
)

type DeleteChat struct {
	chatRepo repository.ChatRepository
	logger   *logger.LogrusLogger
}

func NewDeleteChatUseCase(chatRepo repository.ChatRepository, logger *logger.LogrusLogger) (*DeleteChat, error) {

	return &DeleteChat{chatRepo: chatRepo, logger: logger}, nil
}

func (uc *DeleteChat) DeleteChat(firstID int, secondID int) error {
	uc.logger.Info("DeleteChat", "firstID", firstID, "secondID", secondID)

	err := uc.chatRepo.DeleteChat(firstID, secondID)
	if err != nil {
		uc.logger.Error("DeleteChat", "firstID", firstID, "secondID", secondID, "error", err)
	} else {
		uc.logger.WithFields(&logrus.Fields{"firstID": firstID, "secondID": secondID})
	}
	return err
}
