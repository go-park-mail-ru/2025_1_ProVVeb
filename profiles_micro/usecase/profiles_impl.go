package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	profiles "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/repository"
)

type ProfileServiceServer struct {
	profiles.UnimplementedProfilesServiceServer
	UserRepo   repository.ProfileRepository
	StaticRepo repository.StaticRepository
	Logger     *logger.LogrusLogger
}

func NewProfileServiceServer(
	userRepo repository.ProfileRepository,
	staticRepo repository.StaticRepository,
	logger *logger.LogrusLogger,
) *ProfileServiceServer {
	return &ProfileServiceServer{
		UserRepo:   userRepo,
		StaticRepo: staticRepo,
		Logger:     logger,
	}
}
