package usecase

import (
	"context"

	users "github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/delivery"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/model"
	"github.com/sirupsen/logrus"
)

func (uss *UserServiceServer) SaveUserData(
	ctx context.Context,
	req *users.SaveUserDataRequest,
) (*users.SaveUserDataResponse, error) {
	uss.Logger.Info("StoreUser", "userId", req.UserId)
	var user model.User = model.User{
		UserId:   int(req.UserId),
		Login:    req.User.Login,
		Password: uss.UserRepo.Hash(req.User.Login + "_" + req.User.Password),
		Email:    req.User.Email,
		Phone:    req.User.Phone,
		Status:   int(req.User.Status),
	}
	userId, err := uss.UserRepo.StoreUser(user)
	uss.Logger.WithFields(&logrus.Fields{"userId": req.UserId, "error": err})
	return &users.SaveUserDataResponse{UserId: int32(userId)}, err
}
