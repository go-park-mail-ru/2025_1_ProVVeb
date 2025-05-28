package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	profilespb "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProfileUpdate struct {
	ProfilesService profilespb.ProfilesServiceClient
	logger          *logger.LogrusLogger
}

func NewProfileUpdateUseCase(
	ProfilesService profilespb.ProfilesServiceClient,
	logger *logger.LogrusLogger,
) (*ProfileUpdate, error) {
	if ProfilesService == nil || logger == nil {
		return nil, model.ErrProfileUpdateUC
	}
	return &ProfileUpdate{ProfilesService: ProfilesService, logger: logger}, nil
}

func (pu *ProfileUpdate) UpdateProfile(value model.Profile, targ model.Profile, profileId int) error {
	pu.logger.Info("ProfileUpdateUseCase")

	var valueLikedBy []int32
	for _, likedBy := range value.LikedBy {
		valueLikedBy = append(valueLikedBy, int32(likedBy))
	}
	var valuePrefs []*profilespb.Preference
	for _, pref := range value.Preferences {
		valuePrefs = append(valuePrefs, &profilespb.Preference{
			Description: pref.Description,
			Value:       pref.Value,
		})
	}

	var valueParams []*profilespb.Preference
	for _, pref := range value.Parameters {
		valueParams = append(valueParams, &profilespb.Preference{
			Description: pref.Description,
			Value:       pref.Value,
		})
	}

	var profValue *profilespb.Profile = &profilespb.Profile{
		ProfileId:   int32(value.ProfileId),
		FirstName:   value.FirstName,
		LastName:    value.LastName,
		IsMale:      value.IsMale,
		Height:      int32(value.Height),
		Birthday:    timestamppb.New(value.Birthday),
		Description: value.Description,
		Location:    value.Location,
		Interests:   value.Interests,
		Parametres:  valueParams,
		LikedBy:     valueLikedBy,
		Preferences: valuePrefs,
		Photos:      value.Photos,
	}

	var targLikedBy []int32
	for _, likedBy := range targ.LikedBy {
		targLikedBy = append(targLikedBy, int32(likedBy))
	}
	var targPrefs []*profilespb.Preference
	for _, pref := range targ.Preferences {
		targPrefs = append(targPrefs, &profilespb.Preference{
			Description: pref.Description,
			Value:       pref.Value,
		})
	}

	var targParams []*profilespb.Preference
	for _, pref := range targ.Parameters {
		targParams = append(targParams, &profilespb.Preference{
			Description: pref.Description,
			Value:       pref.Value,
		})
	}

	var targValue *profilespb.Profile = &profilespb.Profile{
		ProfileId:   int32(targ.ProfileId),
		FirstName:   targ.FirstName,
		LastName:    targ.LastName,
		IsMale:      targ.IsMale,
		Goal:        int32(targ.Goal),
		Height:      int32(targ.Height),
		Birthday:    timestamppb.New(targ.Birthday),
		Description: targ.Description,
		Location:    targ.Location,
		Interests:   targ.Interests,
		LikedBy:     targLikedBy,
		Preferences: targPrefs,
		Parametres:  targParams,
		Photos:      targ.Photos,
	}

	req := &profilespb.UpdateProfileRequest{
		Value:     profValue,
		Targ:      targValue,
		ProfileId: int32(profileId),
	}

	_, err := pu.ProfilesService.UpdateProfile(context.Background(), req)

	pu.logger.WithFields(&logrus.Fields{
		"value":     value,
		"targ":      targ,
		"profileId": profileId,
	}).Error("ProfileUpdateUseCase", err)

	return err
}
