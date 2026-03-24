package usecase

import (
	"context"

	users "github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/delivery"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

func (uss *UserServiceServer) GetUserByLogin(
	ctx context.Context,
	req *users.GetUserByLoginRequest,
) (*users.GetUserResponse, error) {
	uss.Logger.Info("GetUser", "login", req.Login)

	user, err := uss.UserRepo.GetUserByLogin(req.Login)
	if err != nil {
		if err == pgx.ErrNoRows {
			uss.Logger.Warn("GetUserByLogin", "login", req.Login, "reason", "user not found")
			return nil, err
		}
		uss.Logger.Error("GetUserByLogin", "login", req.Login, "error", err)
		return nil, err
	}

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
	}).Info("User found")

	return &users.GetUserResponse{
		User: respUser,
	}, nil
}
