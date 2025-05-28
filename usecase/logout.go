package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"

	sessionpb "github.com/go-park-mail-ru/2025_1_ProVVeb/auth_micro/proto"
)

type UserLogOut struct {
	SessionService sessionpb.SessionServiceClient
	logger         *logger.LogrusLogger
}

func NewUserLogOutUseCase(
	SessionService sessionpb.SessionServiceClient,
	logger *logger.LogrusLogger,
) (*UserLogOut, error) {
	if SessionService == nil || logger == nil {
		return nil, model.ErrUserLogOutUC
	}
	return &UserLogOut{
		SessionService: SessionService,
		logger:         logger,
	}, nil
}

func (ul *UserLogOut) Logout(sessionId string) error {
	ul.logger.Info("Logout", "sessionId", sessionId)
	req := &sessionpb.SessionIdRequest{
		SessionId: sessionId,
	}

	sessionResp, err := ul.SessionService.GetSession(context.Background(), req)
	if err != nil {
		ul.logger.Error("Logout", "sessionId", sessionId, "error", err)
		return err
	}

	userIdStr := sessionResp.Data
	req = &sessionpb.SessionIdRequest{
		SessionId: userIdStr,
	}
	ul.logger.Info("Logout", "userId", userIdStr)
	_, err = ul.SessionService.DeleteSession(context.Background(), req)
	if err != nil {
		return err
	}
	ul.logger.Info("Logout", "ok", userIdStr)

	return nil
}
