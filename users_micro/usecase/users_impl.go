package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	users "github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/delivery"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/repository"
)

type UserServiceServer struct {
	users.UnimplementedUsersServiceServer
	UserRepo repository.UserRepository
	Logger   *logger.LogrusLogger
}

func NewUsersServiceServer(
	userRepo repository.UserRepository,
	logger *logger.LogrusLogger,
) *UserServiceServer {
	return &UserServiceServer{
		UserRepo: userRepo,
		Logger:   logger,
	}
}
