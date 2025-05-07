package usecase

import (
	"context"

	profiles "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (pss *ProfileServiceServer) DeleteImage(ctx context.Context, req *profiles.DeleteImageRequest) (*emptypb.Empty, error) {
	pss.Logger.Info("DeleteImage", &logrus.Fields{"user_id": req.GetUserId(), "filename": req.GetFilename()})
	err := pss.StaticRepo.DeleteImage(int(req.GetUserId()), req.GetFilename())
	if err != nil {
		pss.Logger.Error("DeleteImage", &logrus.Fields{"error": err})
		return nil, err
	}
	pss.Logger.Info("DeleteImage: static image deleted")
	err = pss.ProfilesRepo.DeletePhoto(int(req.GetUserId()), req.GetFilename())
	if err != nil {
		pss.Logger.Error("DeleteImage", &logrus.Fields{"error": err})
		return nil, err
	}
	pss.Logger.Info("DeleteImage: user image deleted")
	return &emptypb.Empty{}, nil
}
