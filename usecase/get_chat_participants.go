package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"

	"github.com/sirupsen/logrus"
)

type GetChatParticipants struct {
	chatRepo repository.ChatRepository
	logger   *logger.LogrusLogger
}

func NewGetChatParticipantsUseCase(chatRepo repository.ChatRepository, logger *logger.LogrusLogger) (*GetChatParticipants, error) {

	return &GetChatParticipants{chatRepo: chatRepo, logger: logger}, nil
}

func (uc *GetChatParticipants) GetChatParticipants(chatID int) (int, int, error) {
	uc.logger.Info("GetChatParticipants", "chatID", chatID)
	first, second, err := uc.chatRepo.GetChatParticipants(chatID)
	if err != nil {
		uc.logger.Error("GetProfile", "chatID", chatID, "error", err)
	} else {
		uc.logger.WithFields(&logrus.Fields{"chatID": chatID, "first": first, "second": second})
	}
	return first, second, err
}
