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

	if req.Value.Goal != 0 {
		req.Targ.Goal = req.Value.Goal
	}

	// here is joke:
	// if is male is true, it means that we must change gender
	// sorry bruh
	if req.Value.IsMale {
		req.Targ.IsMale = !req.Targ.IsMale
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

	if len(req.Value.Parametres) != 0 {
		req.Targ.Parametres = req.Value.Parametres
	}

	likedBy := []int{}
	for _, like := range req.Value.LikedBy {
		likedBy = append(likedBy, int(like))
	}

	var prefs []model.Preference
	for _, preference := range req.Value.Preferences {
		prefs = append(prefs, model.Preference{
			Description: preference.Description,
			Value:       preference.Value,
		})
	}

	var params []model.Preference
	for _, preference := range req.Value.Parametres {
		params = append(params, model.Preference{
			Description: preference.Description,
			Value:       preference.Value,
		})
	}

	var prof model.Profile = model.Profile{
		ProfileId:   int(req.ProfileId),
		FirstName:   req.Targ.FirstName,
		IsMale:      req.Targ.IsMale,
		LastName:    req.Targ.LastName,
		Goal:        int(req.Targ.Goal),
		Height:      int(req.Targ.Height),
		Birthday:    req.Targ.Birthday.AsTime(),
		Description: req.Targ.Description,
		Location:    req.Targ.Location,
		Interests:   req.Targ.Interests,
		Preferences: prefs,
		Parameters:  params,
		LikedBy:     likedBy,
	}

	err := pss.ProfilesRepo.UpdateProfile(int(req.ProfileId), prof)
	pss.Logger.WithFields(&logrus.Fields{"error": err}).Error("UpdateProfile")
	return &emptypb.Empty{}, err
}
