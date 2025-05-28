package usecase

import (
	"context"

	profiles "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (pss *ProfileServiceServer) GetProfileStats(
	ctx context.Context,
	req *profiles.GetProfileStatsRequest,
) (*profiles.GetProfileStatsResponse, error) {
	profileID := req.GetProfileId()
	pss.Logger.Info("GetStats", "profile_id", profileID)

	stats, err := pss.ProfilesRepo.GetProfileStats(int(profileID))
	if err != nil {
		pss.Logger.Error("GetStats error", "profile_id", profileID, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get profile stats: %v", err)
	}

	pss.Logger.WithFields(&logrus.Fields{
		"profile_id": profileID,
		"stats":      stats,
	}).Info("Fetched stats")

	return &profiles.GetProfileStatsResponse{
		LikesGiven:         int32(stats.LikesGiven),
		LikesReceived:      int32(stats.LikesReceived),
		Matches:            int32(stats.Matches),
		ComplaintsMade:     int32(stats.ComplaintsMade),
		ComplaintsReceived: int32(stats.ComplaintsReceived),
		MessagesSent:       int32(stats.MessagesSent),
		ChatCount:          int32(stats.ChatCount),
	}, nil
}
