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
		"from":   req.From,
		"to":     req.To,
		"status": req.Status,
	}).Info("SetProfileLike")
	result, err := pss.UserRepo.SetLike(int(req.From), int(req.To), int(req.Status))
	if err != nil {
		pss.Logger.Error("SetProfileLike", "error", err)
	} else {
		pss.Logger.Info("SetProfileLike", "result", result)
	}
	return &profiles.SetProfileLikeResponse{
		LikeId: int32(result),
	}, err
}
