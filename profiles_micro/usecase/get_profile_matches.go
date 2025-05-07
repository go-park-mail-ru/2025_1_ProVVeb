package usecase

import (
	"context"

	profiles "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (pss *ProfileServiceServer) GetProfileMatches(ctx context.Context, req *profiles.GetProfileMatchesRequest) (*profiles.GetProfileMatchesResponse, error) {
	pss.Logger.Info("GetProfileMatches", "user_id", req.GetForUserId())
	result, err := pss.ProfilesRepo.GetMatches(int(req.GetForUserId()))
	if err != nil {
		pss.Logger.WithFields(&logrus.Fields{"forUserId": req.GetForUserId(), "error": err}).Error("GetProfileMatches", "error")
	} else {
		pss.Logger.WithFields(&logrus.Fields{"forUserId": req.GetForUserId(), "dataCount": len(result), "error": err})
	}

	var profs []*profiles.Profile
	for _, profile := range result {
		var prefs []*profiles.Preference
		for _, preference := range profile.Preferences {
			prefs = append(prefs, &profiles.Preference{
				Description: preference.Description,
				Value:       preference.Value,
			})
		}
		profs = append(profs, &profiles.Profile{
			ProfileId:   int32(profile.ProfileId),
			FirstName:   profile.FirstName,
			LastName:    profile.LastName,
			IsMale:      profile.IsMale,
			Height:      int32(profile.Height),
			Birthday:    timestamppb.New(profile.Birthday),
			Description: profile.Description,
			Location:    profile.Location,
			Interests:   profile.Interests,
			// we dont give anybody information about whom profile liked by
			Preferences: prefs,
			Photos:      profile.Photos,
		})
	}

	return &profiles.GetProfileMatchesResponse{Profiles: profs}, err

}
