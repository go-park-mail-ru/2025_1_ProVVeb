package usecase

import (
	"context"

	profiles "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (pss *ProfileServiceServer) DeleteImage(ctx context.Context, req *profiles.DeleteImageRequest) (*emptypb.Empty, error) {
	pss.Logger.Info("DeleteImage", &logrus.Fields{"user_id": req.UserId, "filename": req.Filename})
	err := pss.StaticRepo.DeleteImage(int(req.UserId), req.Filename)
	if err != nil {
		pss.Logger.Error("DeleteImage", &logrus.Fields{"error": err})
		return nil, err
	}
	pss.Logger.Info("DeleteImage: static image deleted")
	err = pss.UserRepo.DeletePhoto(int(req.UserId), req.Filename)
	if err != nil {
		pss.Logger.Error("DeleteImage", &logrus.Fields{"error": err})
		return nil, err
	}
	pss.Logger.Info("DeleteImage: user image deleted")
	return &emptypb.Empty{}, nil
}
