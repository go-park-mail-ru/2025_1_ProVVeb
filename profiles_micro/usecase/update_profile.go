package usecase

import (
	"context"

	profiles "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/model"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (pss *ProfileServiceServer) UpdateProfile(ctx context.Context, req *profiles.UpdateProfileRequest) (*emptypb.Empty, error) {
	pss.Logger.WithFields(&logrus.Fields{"profileId": req.ProfileId, "value profile": req.Value}).Info("UpdateProfile")
	if req.Value.FirstName != "" {
		req.Targ.FirstName = req.Value.FirstName
	}

	if req.Value.LastName != "" {
		req.Targ.LastName = req.Value.LastName
	}

	if req.Value.Height != 0 {
		req.Targ.Height = req.Value.Height
	}

	if req.Value.Birthday.IsValid() {
		req.Targ.Birthday = req.Value.Birthday
	}

	if req.Value.Description != "" {
		req.Targ.Description = req.Value.Description
	}

	if req.Value.Location != "" {
		req.Targ.Location = req.Value.Location
	}

	if len(req.Value.Interests) != 0 {
		req.Targ.Interests = req.Value.Interests
	}

	if len(req.Value.Preferences) != 0 {
		req.Targ.Preferences = req.Value.Preferences
	}

	var prefs []model.Preference
	for _, preference := range req.Value.Preferences {
		prefs = append(prefs, model.Preference{
			Description: preference.Description,
			Value:       preference.Value,
		})
	}

	var prof model.Profile = model.Profile{
		ProfileId:   int(req.ProfileId),
		FirstName:   req.Targ.FirstName,
		LastName:    req.Targ.LastName,
		Height:      int(req.Targ.Height),
		Birthday:    req.Targ.Birthday.AsTime(),
		Description: req.Targ.Description,
		Location:    req.Targ.Location,
		Interests:   req.Targ.Interests,
		Preferences: prefs,
	}

	err := pss.UserRepo.UpdateProfile(int(req.ProfileId), prof)
	pss.Logger.WithFields(&logrus.Fields{"error": err}).Error("UpdateProfile")
	return &emptypb.Empty{}, err
}
