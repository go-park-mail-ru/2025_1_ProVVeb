package usecase

import (
	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	profiles "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/repository"
)

type ProfileServiceServer struct {
	profiles.UnimplementedProfilesServiceServer
	ProfilesRepo repository.ProfileRepository
	StaticRepo   repository.StaticRepository
	Logger       *logger.LogrusLogger
}

func NewProfileServiceServer(
	profilesRepo repository.ProfileRepository,
	staticRepo repository.StaticRepository,
	logger *logger.LogrusLogger,
) *ProfileServiceServer {
	return &ProfileServiceServer{
		ProfilesRepo: profilesRepo,
		StaticRepo:   staticRepo,
		Logger:       logger,
	}
}
