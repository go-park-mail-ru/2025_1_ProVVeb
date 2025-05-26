package usecase

import (
	"context"

	profiles "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (pss *ProfileServiceServer) GetProfile(ctx context.Context, req *profiles.GetProfileRequest) (*profiles.GetProfileResponse, error) {
	pss.Logger.Info("GetProfile", "user_id", req.GetProfileId())
	profile, err := pss.ProfilesRepo.GetProfileById(int(req.GetProfileId()))
	if err != nil {
		pss.Logger.Error("GetProfile", "user_id", req.GetProfileId(), "error", err)
	} else {
		pss.Logger.WithFields(&logrus.Fields{"user_id": req.GetProfileId(), "profile": profile})
	}

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

	var params []*profiles.Preference
	for _, preference := range profile.Parameters {
		params = append(params, &profiles.Preference{
			Description: preference.Description,
			Value:       preference.Value,
		})
	}

	var prof *profiles.Profile = &profiles.Profile{
		ProfileId:   int32(profile.ProfileId),
		FirstName:   profile.FirstName,
		LastName:    profile.LastName,
		IsMale:      profile.IsMale,
		Goal:        int32(profile.Goal),
		Height:      int32(profile.Height),
		Birthday:    timestamppb.New(profile.Birthday),
		Description: profile.Description,
		Location:    profile.Location,
		Interests:   profile.Interests,
		Preferences: prefs,
		Parametres:  params,
		Photos:      profile.Photos,
		LikedBy:     likedBy,
	}
	if profile.Premium.Status {
		prof.Premium = &profiles.Premium{
			Status: profile.Premium.Status,
			Border: int32(profile.Premium.Border),
		}
	}

	return &profiles.GetProfileResponse{Profile: prof}, err

}
