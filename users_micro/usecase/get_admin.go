package usecase

import (
	"context"

	users "github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/delivery"
	"github.com/sirupsen/logrus"
)

func (uss *UserServiceServer) GetAdmin(ctx context.Context, req *users.GetAdminRequest) (*users.GetAdminResponse, error) {
	uss.Logger.Info("GetAdmin", "UserId", req.UserId)
	is_admin, err := uss.UserRepo.GetAdmin(int(req.UserId))
	uss.Logger.WithFields(&logrus.Fields{"UserId": req.UserId, "is_admin": is_admin})
	return &users.GetAdminResponse{IsAdmin: is_admin}, err

}
