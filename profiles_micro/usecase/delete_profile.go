package usecase

import (
	"context"

	profiles "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (pss *ProfileServiceServer) DeleteProfile(
	ctx context.Context,
	req *profiles.DeleteProfileRequest,
) (*emptypb.Empty, error) {
	pss.Logger.WithFields(&logrus.Fields{
		"profileId": req.GetProfileId(),
	}).Info("DeleteProfile")
	err := pss.ProfilesRepo.DeleteProfile(int(req.GetProfileId()))
	pss.Logger.WithFields(&logrus.Fields{
		"veryBigError": err,
	}).Error("DeleteProfile")
	return &emptypb.Empty{}, err
}
