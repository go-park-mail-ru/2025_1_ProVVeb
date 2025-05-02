package usecase

import (
	"context"

	profiles "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (pss *ProfileServiceServer) GetProfile(ctx context.Context, req *profiles.GetProfileRequest) (*profiles.GetProfileResponse, error) {
	pss.Logger.Info("GetProfile", "user_id", req.ProfileId)
	profile, err := pss.UserRepo.GetProfileById(int(req.ProfileId))
	if err != nil {
		pss.Logger.Error("GetProfile", "user_id", req.ProfileId, "error", err)
	} else {
		pss.Logger.WithFields(&logrus.Fields{"user_id": req.ProfileId, "profile": profile})
	}

	var prefs []*profiles.Preference
	for _, preference := range profile.Preferences {
		prefs = append(prefs, &profiles.Preference{
			Description: preference.Description,
			Value:       preference.Value,
		})
	}
	var prof *profiles.Profile = &profiles.Profile{
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
	}

	return &profiles.GetProfileResponse{Profile: prof}, err

}
