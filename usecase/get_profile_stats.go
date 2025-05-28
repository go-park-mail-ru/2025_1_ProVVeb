package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	profilespb "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/sirupsen/logrus"
)

type GetProfileStats struct {
	ProfilesService profilespb.ProfilesServiceClient
	logger          *logger.LogrusLogger
}

func NewGetProfileStatsUseCase(
	ProfilesService profilespb.ProfilesServiceClient,
	logger *logger.LogrusLogger,
) (*GetProfileStats, error) {
	if ProfilesService == nil || logger == nil {
		return &GetProfileStats{}, model.ErrGetProfileUC
	}
	return &GetProfileStats{ProfilesService: ProfilesService, logger: logger}, nil
}

func (gp *GetProfileStats) GetProfileStats(userId int) (model.ProfileStats, error) {
	gp.logger.Info("GetProfileStats use case called")

	req := &profilespb.GetProfileStatsRequest{
		ProfileId: int32(userId),
	}

	res, err := gp.ProfilesService.GetProfileStats(context.Background(), req)
	if err != nil {
		gp.logger.WithFields(&logrus.Fields{
			"error": err,
		}).Error("GetProfileStats failed")
		return model.ProfileStats{}, err
	}

	stats := model.ProfileStats{
		LikesGiven:         int(res.LikesGiven),
		LikesReceived:      int(res.LikesReceived),
		Matches:            int(res.Matches),
		ComplaintsMade:     int(res.ComplaintsMade),
		ComplaintsReceived: int(res.ComplaintsReceived),
		MessagesSent:       int(res.MessagesSent),
		ChatCount:          int(res.ChatCount),
	}

	gp.logger.WithFields(&logrus.Fields{
		"profile_stats": stats,
	}).Info("GetProfileStats succeeded")

	return stats, nil
}
