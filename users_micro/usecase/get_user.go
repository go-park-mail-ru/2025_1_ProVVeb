package usecase

import (
	"context"

	users "github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/delivery"
	"github.com/sirupsen/logrus"
)

func (uss *UserServiceServer) GetUser(
	ctx context.Context,
	req *users.GetUserRequest,
) (*users.GetUserResponse, error) {
	uss.Logger.Info("GetUser", "userId", req.UserId)
	user, err := uss.UserRepo.GetUserParams(int(req.UserId))
	uss.Logger.WithFields(&logrus.Fields{
		"userId": req.UserId,
		"error":  err,
	})

	respUser := &users.User{
		UserId:   int32(user.UserId),
		Login:    user.Login,
		Password: user.Password,
		Email:    user.Email,
		Phone:    user.Phone,
		Status:   int32(user.Status),
	}
	uss.Logger.WithFields(&logrus.Fields{
		"userId": respUser.UserId,
		"login":  respUser.Login,
		"email":  respUser.Email,
		"phone":  respUser.Phone,
		"status": respUser.Status,
	})
	return &users.GetUserResponse{
		User: respUser,
	}, err
}
