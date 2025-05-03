package usecase

import (
	"context"

	users "github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/delivery"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (uss *UserServiceServer) DeleteUser(
	ctx context.Context,
	req *users.DeleteUserRequest,
) (*emptypb.Empty, error) {
	uss.Logger.Info("DeleteUser", "userId", req.UserId)
	err := uss.UserRepo.DeleteUserById(int(req.UserId))
	uss.Logger.WithFields(&logrus.Fields{"userId": req.UserId, "error": err})
	return &emptypb.Empty{}, err
}
