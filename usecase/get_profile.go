package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	profilespb "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/sirupsen/logrus"
)

type GetProfile struct {
	ProfilesService profilespb.ProfilesServiceClient
	logger          *logger.LogrusLogger
}

func NewGetProfileUseCase(
	ProfilesService profilespb.ProfilesServiceClient,
	logger *logger.LogrusLogger,
) (*GetProfile, error) {
	if ProfilesService == nil || logger == nil {
		return &GetProfile{}, model.ErrGetProfileUC
	}
	return &GetProfile{ProfilesService: ProfilesService, logger: logger}, nil
}

func (gp *GetProfile) GetProfile(userId int) (model.Profile, error) {
	gp.logger.Info("GetProfileUseCase")
	req := &profilespb.GetProfileRequest{
		ProfileId: int32(userId),
	}
	res, err := gp.ProfilesService.GetProfile(context.Background(), req)
	if err != nil {
		gp.logger.WithFields(&logrus.Fields{
			"error": err,
		}).Error("GetProfileUseCase")
		return model.Profile{}, err
	}

	var prefs []model.Preference
	for _, pref := range res.Profile.Preferences {
		prefs = append(prefs, model.Preference{
			Description: pref.Description,
			Value:       pref.Value,
		})
	}
	var likedBy []int
	for _, like := range res.Profile.LikedBy {
		likedBy = append(likedBy, int(like))
	}

	var profile model.Profile = model.Profile{
		ProfileId:   int(res.Profile.ProfileId),
		FirstName:   res.Profile.FirstName,
		LastName:    res.Profile.LastName,
		IsMale:      res.Profile.IsMale,
		Height:      int(res.Profile.Height),
		Birthday:    res.Profile.Birthday.AsTime(),
		Description: res.Profile.Description,
		Location:    res.Profile.Location,
		Interests:   res.Profile.Interests,
		LikedBy:     likedBy,
		Preferences: prefs,
		Photos:      res.Profile.Photos,
	}

	gp.logger.WithFields(&logrus.Fields{
		"profile": profile,
	}).Info("GetProfileUseCase")
	return profile, err

}
