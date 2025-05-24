package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	profilespb "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
)

type SearchProfiles struct {
	ProfilesService profilespb.ProfilesServiceClient
	logger          *logger.LogrusLogger
}

func NewSearchProfilesUseCase(
	ProfilesService profilespb.ProfilesServiceClient,
	logger *logger.LogrusLogger,
) (*SearchProfiles, error) {
	if ProfilesService == nil || logger == nil {
		return nil, model.ErrGetProfilesForUserUC
	}

	return &SearchProfiles{
		ProfilesService: ProfilesService,
		logger:          logger,
	}, nil
}
func (gp *SearchProfiles) GetSearchProfiles(forUserId int, params model.SearchProfileRequest) ([]model.FoundProfile, error) {
	gp.logger.Info("GetProfilesForUserUseCase")

	req := &profilespb.SearchProfileRequest{
		IDUser:     int32(forUserId),
		Input:      params.Input,
		IsMale:     params.IsMale,
		AgeMin:     int32(params.AgeMin),
		AgeMax:     int32(params.AgeMax),
		HeightMin:  int32(params.HeightMin),
		HeightMax:  int32(params.HeightMax),
		Goal:       int32(params.Goal),
		Parametres: make([]*profilespb.Preference, len(params.Preferences)), // нужно преобразовать
	}

	for i, p := range params.Preferences {
		req.Parametres[i] = &profilespb.Preference{
			Description: p.Description,
			Value:       p.Value,
		}
	}

	resp, err := gp.ProfilesService.SearchProfile(context.Background(), req)
	if err != nil {
		return nil, err
	}

	var profs []model.FoundProfile

	for _, match := range resp.Profiles {
		if int(match.IDUser) == forUserId {
			continue
		}

		profs = append(profs, model.FoundProfile{
			IDUser:   int(match.IDUser),
			FirstImg: match.FirstImg,
			Fullname: match.Fullname,
			Age:      int(match.Age),
			Goal:     int(match.Goal),
		})
	}

	return profs, nil
}
