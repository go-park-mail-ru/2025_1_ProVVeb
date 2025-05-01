package usecase

import (
	"context"
	"strconv"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"

	sessionpb "github.com/go-park-mail-ru/2025_1_ProVVeb/auth_micro/proto"
	"github.com/sirupsen/logrus"
)

type UserCheckSession struct {
	SessionService sessionpb.SessionServiceClient
	logger         *logger.LogrusLogger
}

func NewUserCheckSessionUseCase(SessionService sessionpb.SessionServiceClient, logger *logger.LogrusLogger) (*UserCheckSession, error) {

	return &UserCheckSession{SessionService: SessionService, logger: logger}, nil
}

func (uc *UserCheckSession) CheckSession(sessionId string) (int, error) {
	req := &sessionpb.SessionIdRequest{
		SessionId: sessionId,
	}

	sessionResp, err := uc.SessionService.GetSession(context.Background(), req)
	if err != nil {
		uc.logger.Error("Session not found")
		return -1, err
	}
	uc.logger.Info("Session found")

	userIdStr := sessionResp.Data

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		uc.logger.Error("Invalid session id")
		return -1, model.ErrInvalidSessionId
	}
	uc.logger.WithFields(&logrus.Fields{"userId": userId}).Info("Session checked")
	return userId, nil
}
