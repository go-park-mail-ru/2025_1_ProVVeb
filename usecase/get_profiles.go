package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	profilespb "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/sirupsen/logrus"
)

type GetProfilesForUser struct {
	ProfilesService profilespb.ProfilesServiceClient
	logger          *logger.LogrusLogger
}

func NewGetProfilesForUserUseCase(
	ProfilesService profilespb.ProfilesServiceClient,
	logger *logger.LogrusLogger,
) (*GetProfilesForUser, error) {
	if ProfilesService == nil || logger == nil {
		return nil, model.ErrGetProfilesForUserUC
	}

	return &GetProfilesForUser{
		ProfilesService: ProfilesService,
		logger:          logger,
	}, nil
}

func (gp *GetProfilesForUser) GetProfiles(forUserId int) ([]model.Profile, error) {
	gp.logger.Info("GetProfilesForUserUseCase")
	req := &profilespb.GetProfilesRequest{
		ForUserId: int32(forUserId),
	}
	resp, err := gp.ProfilesService.GetProfiles(context.Background(), req)

	var profs []model.Profile
	for _, match := range resp.Profiles {
		var prefs []model.Preference
		for _, pref := range match.Preferences {
			prefs = append(prefs, model.Preference{
				Description: pref.Description,
				Value:       pref.Value,
			})
		}
		var params []model.Preference
		for _, pref := range match.Parametres {
			params = append(params, model.Preference{
				Description: pref.Description,
				Value:       pref.Value,
			})
		}
		profs = append(profs, model.Profile{
			ProfileId:   int(match.ProfileId),
			FirstName:   match.FirstName,
			LastName:    match.LastName,
			IsMale:      match.IsMale,
			Goal:        int(match.Goal),
			Height:      int(match.Height),
			Birthday:    match.Birthday.AsTime(),
			Description: match.Description,
			Location:    match.Location,
			Interests:   match.Interests,
			// LikedBy:	 match.LikedBy,
			// give smbd information about by whom the user is liked?..
			Preferences: prefs,
			Parameters:  params,
			Photos:      match.Photos,
			Premium: model.Premium{
				Status: match.Premium.Status,
				Border: int32(match.Premium.Border),
			},
		})
	}
	gp.logger.WithFields(&logrus.Fields{
		"len matches": len(profs),
		"method":      "GetProfileMatches",
		"error":       err,
	})
	return profs, err

}
