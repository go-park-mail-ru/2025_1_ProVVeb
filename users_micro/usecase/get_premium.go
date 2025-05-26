package usecase

import (
	"context"
	"time"

	users "github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/delivery"
)

func (uss *UserServiceServer) GetPremium(ctx context.Context, req *users.GetPremiumRequest) (*users.GetPremiumResponse, error) {
	uss.Logger.Info("GetPremium", "UserId", req.UserId)

	hasSub, subType, expiresAt, err := uss.UserRepo.GetPremium(int(req.UserId))

	if !hasSub || expiresAt == nil || time.Now().After(*expiresAt) {
		return &users.GetPremiumResponse{
			UserId:     req.UserId,
			IsSubsribe: false,
			Type:       0,
		}, err
	}

	return &users.GetPremiumResponse{
		UserId:     req.UserId,
		IsSubsribe: true,
		Type:       int32(subType),
	}, err
}
