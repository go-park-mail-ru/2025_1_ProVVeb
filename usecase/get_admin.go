package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	userspb "github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/delivery"
	"github.com/sirupsen/logrus"
)

type GetAdmin struct {
	UsersService userspb.UsersServiceClient
	logger       *logger.LogrusLogger
}

func NewGetAdminUseCase(
	UsersService userspb.UsersServiceClient,
	logger *logger.LogrusLogger,
) (*GetAdmin, error) {
	if UsersService == nil || logger == nil {
		return nil, model.ErrUserDeleteUC
	}
	return &GetAdmin{
		UsersService: UsersService,
		logger:       logger,
	}, nil
}

func (ga *GetAdmin) GetAdmin(userId int) (bool, error) {
	ga.logger.Info("GetAdmin", "userId", userId)
	userReq := &userspb.GetAdminRequest{
		UserId: int32(userId),
	}
	is_admin, err := ga.UsersService.GetAdmin(context.Background(), userReq)
	ga.logger.WithFields(&logrus.Fields{"GetAdmin": userId}).Info("DeleteUser")

	return is_admin.IsAdmin, err
}
