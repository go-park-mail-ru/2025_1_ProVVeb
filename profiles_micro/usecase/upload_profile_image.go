package usecase

import (
	"context"

	profiles "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (pss *ProfileServiceServer) UploadProfileImage(
	ctx context.Context,
	req *profiles.UploadProfileImageRequest,
) (*emptypb.Empty, error) {
	pss.Logger.Info("UploadProfileImage")
	filenameWithPath := "/" + req.GetFilename()
	err := pss.StaticRepo.UploadImage(req.GetFile(), filenameWithPath, req.GetContentType())
	pss.Logger.WithFields(&logrus.Fields{
		"user_id":      req.GetUserId(),
		"filename":     req.GetFilename(),
		"content_type": req.GetContentType(),
	})
	if err != nil {
		pss.Logger.Error("UploadProfileImage", err)
		return nil, err
	}
	pss.Logger.Info("UploadProfileImage: image uploaded")
	err = pss.ProfilesRepo.StorePhoto(int(req.GetUserId()), req.GetFilename())
	pss.Logger.Info("error", err)
	return &emptypb.Empty{}, err
}
