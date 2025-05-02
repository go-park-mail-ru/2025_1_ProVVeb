package usecase

import (
	"context"

	profiles "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/sirupsen/logrus"
)

func (pss *ProfileServiceServer) GetProfileImages(ctx context.Context, req *profiles.GetProfileImagesRequest) (*profiles.GetProfileImagesResponse, error) {
	pss.Logger.Info("GetProfileImages", "user_id", req.UserId)
	urls, err := pss.UserRepo.GetPhotos(int(req.UserId))
	if err != nil {
		pss.Logger.Error("GetProfileImages", "user_id", req.UserId, "urls error", err)
		return nil, err
	}
	pss.Logger.Info("GetProfileImages", "user_id", req.UserId, "urls", urls)
	files, err := pss.StaticRepo.GetImages(urls)
	if err != nil {
		pss.Logger.Error("GetProfileImages", "user_id", req.UserId, "files error", err)
		return nil, err
	}
	pss.Logger.WithFields(&logrus.Fields{"user_id": req.UserId, "urlsCount": len(urls)}).Info("ok")
	return &profiles.GetProfileImagesResponse{Files: files, Urls: urls}, nil
}
