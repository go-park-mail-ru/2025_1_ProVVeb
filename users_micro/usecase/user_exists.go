package usecase

import (
	"context"

	users "github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/delivery"
	"github.com/sirupsen/logrus"
)

func (uss *UserServiceServer) UserExists(
	ctx context.Context,
	req *users.UserExistsRequest,
) (*users.UserExistsResponse, error) {
	uss.Logger.Info("UserExists", "login", req.Login)
	exists := uss.UserRepo.UserExists(req.Login)
	uss.Logger.WithFields(&logrus.Fields{"login": req.Login, "exists": exists})
	return &users.UserExistsResponse{Exists: exists}, nil
}