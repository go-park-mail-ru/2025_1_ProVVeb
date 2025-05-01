package usecase

import (
	"context"
	"strconv"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"

	sessionpb "github.com/go-park-mail-ru/2025_1_ProVVeb/auth_micro/proto"
)

type UserCheckSession struct {
	SessionService sessionpb.SessionServiceClient
}

func NewUserCheckSessionUseCase(SessionService sessionpb.SessionServiceClient) (*UserCheckSession, error) {

	return &UserCheckSession{SessionService: SessionService}, nil
}

func (uc *UserCheckSession) CheckSession(sessionId string) (int, error) {
	req := &sessionpb.SessionIdRequest{
		SessionId: sessionId,
	}

	sessionResp, err := uc.SessionService.GetSession(context.Background(), req)
	if err != nil {
		return -1, err
	}

	userIdStr := sessionResp.Data

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return -1, model.ErrInvalidSessionId
	}
	return userId, nil
}
