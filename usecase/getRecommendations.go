package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	profilespb "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/sirupsen/logrus"
)

type GetRecommendations struct {
	ProfilesService profilespb.ProfilesServiceClient
	logger          *logger.LogrusLogger
}

func NewGetRecommendationsUseCase(
	ProfilesService profilespb.ProfilesServiceClient,
	logger *logger.LogrusLogger,
) (*GetRecommendations, error) {
	if ProfilesService == nil || logger == nil {
		return &GetRecommendations{}, model.ErrGetProfileUC
	}
	return &GetRecommendations{ProfilesService: ProfilesService, logger: logger}, nil
}

func (gp *GetRecommendations) GetRecommendations(userId int) (model.Profile, error) {
	gp.logger.Info("GetRecommendations")
	req := &profilespb.GetProfileRequest{
		ProfileId: int32(userId),
	}
	res, err := gp.ProfilesService.GetRecommendations(context.Background(), req)
	if err != nil {
		gp.logger.WithFields(&logrus.Fields{
			"error": err,
		}).Error("GetRecommendationsUseCase")
		return model.Profile{}, err
	}

	var prefs []model.Preference
	for _, pref := range res.Profile.Preferences {
		prefs = append(prefs, model.Preference{
			Description: pref.Description,
			Value:       pref.Value,
		})
	}

	var params []model.Preference
	for _, pref := range res.Profile.Parametres {
		params = append(params, model.Preference{
			Description: pref.Description,
			Value:       pref.Value,
		})
	}

	var likedBy []int
	for _, like := range res.Profile.LikedBy {
		likedBy = append(likedBy, int(like))
	}

	premium := model.Premium{}
	if res.Profile.Premium != nil {
		premium.Status = res.Profile.Premium.Status
		premium.Border = int32(res.Profile.Premium.Border)
	}

	var profile model.Profile = model.Profile{
		ProfileId:   int(res.Profile.ProfileId),
		FirstName:   res.Profile.FirstName,
		LastName:    res.Profile.LastName,
		IsMale:      res.Profile.IsMale,
		Height:      int(res.Profile.Height),
		Goal:        int(res.Profile.Goal),
		Birthday:    res.Profile.Birthday.AsTime(),
		Description: res.Profile.Description,
		Location:    res.Profile.Location,
		Interests:   res.Profile.Interests,
		LikedBy:     likedBy,
		Preferences: prefs,
		Parameters:  params,
		Photos:      res.Profile.Photos,
		Premium:     premium,
	}

	gp.logger.WithFields(&logrus.Fields{
		"profile": profile,
	}).Info("GetProfileUseCase")
	return profile, err

}
