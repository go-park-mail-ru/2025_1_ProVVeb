package usecase

import (
	"context"

	profiles "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (pss *ProfileServiceServer) GetProfiles(
	ctx context.Context,
	req *profiles.GetProfilesRequest,
) (*profiles.GetProfilesResponse, error) {
	pss.Logger.Info("GetProfiles", "forUserId", req.GetForUserId())
	result, err := pss.ProfilesRepo.GetProfilesByUserId(int(req.GetForUserId()))
	if err != nil {
		pss.Logger.Error("GetProfiles", "forUserId", req.GetForUserId(), "error", err)
	} else {
		pss.Logger.WithFields(&logrus.Fields{
			"forUserId":     req.GetForUserId(),
			"profilesCount": len(result),
		}).Info("GetProfiles")
	}

	var profs []*profiles.Profile
	for _, profile := range result {
		likedBy := []int32{}
		for _, like := range profile.LikedBy {
			likedBy = append(likedBy, int32(like))
		}
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
			Preferences: prefs,
			Photos:      profile.Photos,
			LikedBy:     likedBy,
		})
	}

	return &profiles.GetProfilesResponse{Profiles: profs}, nil
}
