package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	profilespb "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/sirupsen/logrus"
)

type GetProfileMatches struct {
	ProfilesService profilespb.ProfilesServiceClient
	logger          *logger.LogrusLogger
}

func NewGetProfileMatchesUseCase(
	ProfilesService profilespb.ProfilesServiceClient,
	logger *logger.LogrusLogger,
) (*GetProfileMatches, error) {
	if ProfilesService == nil || logger == nil {
		return nil, model.ErrGetProfileMatchesUC
	}
	return &GetProfileMatches{ProfilesService: ProfilesService, logger: logger}, nil
}

func (gp *GetProfileMatches) GetMatches(forUserId int) ([]model.Profile, error) {
	gp.logger.WithFields(&logrus.Fields{"forUserId": forUserId, "method": "GetProfileMatches"})
	req := &profilespb.GetProfileMatchesRequest{
		ForUserId: int32(forUserId),
	}
	resp, err := gp.ProfilesService.GetProfileMatches(context.Background(), req)

	var matches []model.Profile
	for _, match := range resp.Profiles {
		var prefs []model.Preference
		for _, pref := range match.Preferences {
			prefs = append(prefs, model.Preference{
				Description: pref.Description,
				Value:       pref.Value,
			})
		}
		matches = append(matches, model.Profile{
			ProfileId:   int(match.ProfileId),
			FirstName:   match.FirstName,
			LastName:    match.LastName,
			IsMale:      match.IsMale,
			Height:      int(match.Height),
			Birthday:    match.Birthday.AsTime(),
			Description: match.Description,
			Location:    match.Location,
			Interests:   match.Interests,
			// LikedBy:	 match.LikedBy,
			// give smbd information about by whom the user is liked?..
			Preferences: prefs,
			Photos:      match.Photos,
		})
	}
	gp.logger.WithFields(&logrus.Fields{
		"len matches": len(matches),
		"method":      "GetProfileMatches",
		"error":       err,
	})
	return matches, err
}
