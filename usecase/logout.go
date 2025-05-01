package usecase

import (
	"context"
	"strconv"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"

	sessionpb "github.com/go-park-mail-ru/2025_1_ProVVeb/auth_micro/proto"
)

type UserLogOut struct {
	userRepo       repository.UserRepository
	SessionService sessionpb.SessionServiceClient
}

func NewUserLogOutUseCase(
	userRepo repository.UserRepository,
	SessionService sessionpb.SessionServiceClient,
) (*UserLogOut, error) {
	if userRepo == nil {
		return nil, model.ErrUserLogOutUC
	}
	return &UserLogOut{
		userRepo:       userRepo,
		SessionService: SessionService,
	}, nil
}

func (ul *UserLogOut) Logout(sessionId string) error {
	req := &sessionpb.SessionIdRequest{
		SessionId: sessionId,
	}

	sessionResp, err := ul.SessionService.GetSession(context.Background(), req)
	if err != nil {
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

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return model.ErrInvalidSessionId
	}

	err = ul.userRepo.DeleteSession(userId)
	if err != nil {
		return model.ErrDeleteSession
	}

	return nil
}

//
