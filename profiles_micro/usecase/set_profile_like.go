package usecase

import (
	"context"

	profiles "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/sirupsen/logrus"
)

func (pss *ProfileServiceServer) SetProfileLike(
	ctx context.Context,
	req *profiles.SetProfileLikeRequest,
) (*profiles.SetProfileLikeResponse, error) {
	pss.Logger.WithFields(&logrus.Fields{
		"from":   req.GetFrom(),
		"to":     req.GetTo(),
		"status": req.GetStatus(),
	}).Info("SetProfileLike")
	result, err := pss.ProfilesRepo.SetLike(
		int(req.GetFrom()),
		int(req.GetTo()),
		int(req.GetStatus()),
	)
	if err != nil {
		pss.Logger.Error("SetProfileLike", "error", err)
	} else {
		pss.Logger.Info("SetProfileLike", "result", result)
	}
	return &profiles.SetProfileLikeResponse{
		LikeId: int32(result),
	}, err
}
