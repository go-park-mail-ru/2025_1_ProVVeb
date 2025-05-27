package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	userspb "github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/delivery"
	"github.com/sirupsen/logrus"
)

type UserGetParams struct {
	UsersService userspb.UsersServiceClient
	logger       *logger.LogrusLogger
}

func NewUserGetParamsUseCase(
	UsersService userspb.UsersServiceClient,
	logger *logger.LogrusLogger,
) (*UserGetParams, error) {
	if UsersService == nil || logger == nil {
		return nil, model.ErrUserGetParamsUC
	}
	return &UserGetParams{UsersService: UsersService, logger: logger}, nil
}

func (up *UserGetParams) GetUserParams(userId int) (model.User, error) {
	up.logger.Info("GetUserParams", "userId", userId)
	req := &userspb.GetUserRequest{UserId: int32(userId)}
	res, err := up.UsersService.GetUser(context.Background(), req)
	if err != nil {
		up.logger.Error("GetUserParams", "error", err)
		return model.User{}, err
	}
	user := model.User{
		UserId:   int(res.User.UserId),
		Login:    res.User.Login,
		Password: res.User.Password,
		Email:    res.User.Email,
		Phone:    res.User.Phone,
		Status:   int(res.User.Status),
	}
	up.logger.WithFields(&logrus.Fields{
		"userId": user.UserId,
		"login":  user.Login,
		"email":  user.Email,
		"phone":  user.Phone,
		"status": user.Status,
		"error":  err,
	})
	return user, err
}
