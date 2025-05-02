package usecase

import (
	"context"

	profiles "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (pss *ProfileServiceServer) UploadProfileImage(ctx context.Context, req *profiles.UploadProfileImageRequest) (*emptypb.Empty, error) {
	pss.Logger.Info("UploadProfileImage")
	filenameWithPath := "/" + req.Filename
	err := pss.StaticRepo.UploadImage(req.File, filenameWithPath, req.ContentType)
	pss.Logger.WithFields(&logrus.Fields{"user_id": req.UserId, "filename": req.Filename, "content_type": req.ContentType})
	if err != nil {
		pss.Logger.Error("UploadProfileImage", err)
		return nil, err
	}
	pss.Logger.Info("UploadProfileImage: image uploaded")
	err = pss.UserRepo.StorePhoto(int(req.UserId), req.Filename)
	pss.Logger.Info("error", err)
	return &emptypb.Empty{}, err
}
