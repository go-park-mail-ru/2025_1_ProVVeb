package usecase

import (
	"context"

	users "github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/delivery"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (uss *UserServiceServer) ValidateLogin(
	ctx context.Context,
	req *users.ValidateLoginRequest,
) (*emptypb.Empty, error) {
	uss.Logger.Info("ValidateLogin", "login", req.Login)
	err := uss.UserRepo.ValidateLogin(req.Login)
	uss.Logger.WithFields(&logrus.Fields{"login": req.Login, "error": err})
	return &emptypb.Empty{}, err
}

func (uss *UserServiceServer) ValidatePassword(
	ctx context.Context,
	req *users.ValidatePasswordRequest,
) (*emptypb.Empty, error) {
	uss.Logger.Info("ValidatePassword", "password", req.Password)
	err := uss.UserRepo.ValidatePassword(req.Password)
	uss.Logger.WithFields(&logrus.Fields{"password": req.Password, "error": err})
	return &emptypb.Empty{}, err
}
