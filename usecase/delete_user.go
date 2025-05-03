package usecase

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	profilespb "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	userspb "github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/delivery"
	"github.com/sirupsen/logrus"
)

type DeleteUser struct {
	UsersService    userspb.UsersServiceClient
	ProfilesService profilespb.ProfilesServiceClient
	logger          *logger.LogrusLogger
}

func NewUserDeleteUseCase(
	UsersService userspb.UsersServiceClient,
	ProfilesService profilespb.ProfilesServiceClient,
	logger *logger.LogrusLogger,
) (*DeleteUser, error) {
	if UsersService == nil || ProfilesService == nil || logger == nil {
		return nil, model.ErrUserDeleteUC
	}
	return &DeleteUser{
		UsersService:    UsersService,
		ProfilesService: ProfilesService,
		logger:          logger,
	}, nil
}

func (du *DeleteUser) DeleteUser(userId int) error {
	fmt.Println("1")
	du.logger.Info("DeleteUser", "userId", userId)
	userReq := &userspb.DeleteUserRequest{
		UserId: int32(userId),
	}
	fmt.Println("2")
	_, err := du.UsersService.DeleteUser(context.Background(), userReq)
	if err != nil {
		du.logger.WithFields(&logrus.Fields{"userId": userId, "error": err}).Error("DeleteUser")
		return err
	}
	du.logger.WithFields(&logrus.Fields{"userId": userId}).Info("DeleteUser")

	fmt.Println("3")
	du.logger.Info("DeleteProfile")
	profileReq := &profilespb.DeleteProfileRequest{
		ProfileId: int32(userId),
	}

	fmt.Println("4")
	_, err = du.ProfilesService.DeleteProfile(context.Background(), profileReq)
	du.logger.WithFields(&logrus.Fields{"profileId": userId, "error": err}).Info("DeleteProfile")
	fmt.Println("5")

	return err
}
