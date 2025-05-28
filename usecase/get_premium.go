package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	userspb "github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/delivery"
	"github.com/sirupsen/logrus"
)

type GetPremium struct {
	UsersService userspb.UsersServiceClient
	logger       *logger.LogrusLogger
}

func NewGetPremiumUseCase(
	UsersService userspb.UsersServiceClient,
	logger *logger.LogrusLogger,
) (*GetPremium, error) {
	if UsersService == nil || logger == nil {
		return nil, model.ErrUserDeleteUC
	}
	return &GetPremium{
		UsersService: UsersService,
		logger:       logger,
	}, nil
}

func (ga *GetPremium) GetPremium(userId int) (bool, int, error) {
	ga.logger.Info("GetAdmin", "userId", userId)
	userReq := &userspb.GetPremiumRequest{
		UserId: int32(userId),
	}
	is_premium, err := ga.UsersService.GetPremium(context.Background(), userReq)
	if err != nil {
		ga.logger.WithFields(&logrus.Fields{
			"GetAdminError": userId,
			"error":         err.Error(),
		}).Error("UsersService.GetPremium failed")

		return false, 0, err
	}
	ga.logger.WithFields(&logrus.Fields{"GetAdmin": userId}).Info("DeleteUser")

	return is_premium.IsSubsribe, int(is_premium.Type), err
}
