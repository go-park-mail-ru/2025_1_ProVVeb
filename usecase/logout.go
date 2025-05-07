package usecase

import (
	"context"
	"strconv"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"

	sessionpb "github.com/go-park-mail-ru/2025_1_ProVVeb/auth_micro/proto"
)

type UserLogOut struct {
	userRepo       repository.UserRepository
	SessionService sessionpb.SessionServiceClient
	logger         *logger.LogrusLogger
}

func NewUserLogOutUseCase(
	userRepo repository.UserRepository,
	SessionService sessionpb.SessionServiceClient,
	logger *logger.LogrusLogger,
) (*UserLogOut, error) {
	if userRepo == nil || SessionService == nil || logger == nil {
		return nil, model.ErrUserLogOutUC
	}
	return &UserLogOut{
		userRepo:       userRepo,
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
	_, err = ul.SessionService.DeleteSession(context.Background(), req)
	if err != nil {
		return err
	}

	ul.logger.Info("Logout", "userId", userIdStr)
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		ul.logger.Error("Logout", "userId", userIdStr, "error", err)
		return model.ErrInvalidSessionId
	}

	ul.logger.Info("Logout", "userId", userId)
	err = ul.userRepo.DeleteSession(userId)
	if err != nil {
		ul.logger.Error("Logout", "userId", userId, "error", err)
		return model.ErrDeleteSession
	}
	ul.logger.Info("Logout", "ok", userId)

	return nil
}
