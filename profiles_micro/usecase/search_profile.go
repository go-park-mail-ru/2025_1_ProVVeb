package usecase

import (
	"context"

	profiles "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/model"
)

func (pss *ProfileServiceServer) SearchProfile(
	ctx context.Context,
	req *profiles.SearchProfileRequest,
) (*profiles.SearchProfileResponse, error) {

	params := model.SearchProfileRequest{
		IsMale:      req.GetIsMale(),
		Input:       req.GetInput(),
		AgeMin:      int(req.GetAgeMin()),
		AgeMax:      int(req.GetAgeMax()),
		HeightMin:   int(req.GetHeightMin()),
		HeightMax:   int(req.GetHeightMax()),
		Goal:        int(req.GetGoal()),
		Country:     req.GetCountry(),
		City:        req.GetCity(),
		Preferences: make([]model.Preference, 0, len(req.GetParametres())),
	}
	for _, p := range req.GetParametres() {
		params.Preferences = append(params.Preferences, model.Preference{
			Description: p.GetDescription(),
			Value:       p.GetValue(),
		})
	}

	foundProfiles, err := pss.ProfilesRepo.SearchProfiles(int(req.GetIDUser()), params)
	if err != nil {
		return nil, err
	}

	resp := &profiles.SearchProfileResponse{}

	for _, p := range foundProfiles {
		resp.Profiles = append(resp.Profiles, &profiles.FoundProfile{
			IDUser:   int32(p.IDUser),
			FirstImg: p.FirstImg,
			Fullname: p.Fullname,
			Age:      int32(p.Age),
			Goal:     int32(p.Goal),
		})
	}

	return resp, nil
}
