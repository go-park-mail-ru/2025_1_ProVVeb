package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	profilespb "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
)

type ProfileSetLike struct {
	ProfileService profilespb.ProfilesServiceClient
	logger         *logger.LogrusLogger
}

func NewProfileSetLikeUseCase(
	ProfileService profilespb.ProfilesServiceClient,
	logger *logger.LogrusLogger,
) (*ProfileSetLike, error) {
	if ProfileService == nil || logger == nil {
		return nil, model.ErrProfileSetLikeUC
	}
	return &ProfileSetLike{ProfileService: ProfileService, logger: logger}, nil
}

func (l *ProfileSetLike) SetLike(from int, to int, status int) (int, error) {
	l.logger.Info("ProfileSetLikeUseCase")
	req := &profilespb.SetProfileLikeRequest{
		From:   int32(from),
		To:     int32(to),
		Status: int32(status),
	}
	resp, err := l.ProfileService.SetProfileLike(context.Background(), req)
	if err != nil {
		l.logger.Error("ProfileSetLikeUseCase", err)
		return 0, err
	}
	return int(resp.LikeId), nil
}
