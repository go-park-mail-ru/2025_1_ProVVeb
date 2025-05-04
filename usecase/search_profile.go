package usecase

import (
	"context"
	"strings"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	profilespb "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/utils"
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

	req := &profilespb.GetProfilesRequest{
		ForUserId: int32(forUserId),
	}
	resp, err := gp.ProfilesService.GetProfiles(context.Background(), req)
	if err != nil {
		return nil, err
	}

	var profs []model.FoundProfile

	for _, match := range resp.Profiles {
		if int(match.ProfileId) == forUserId {
			continue
		}

		if params.IsMale != "Any" {
			if (params.IsMale == "Male" && !match.IsMale) || (params.IsMale == "Female" && match.IsMale) {
				continue
			}
		}

		age := utils.CalculateAge(match.Birthday.AsTime())
		if params.AgeMin > 0 && age < params.AgeMin {
			continue
		}
		if params.AgeMax > 0 && age > params.AgeMax {
			continue
		}

		if params.HeightMin > 0 && int(match.Height) < params.HeightMin {
			continue
		}
		if params.HeightMax > 0 && int(match.Height) > params.HeightMax {
			continue
		}

		locParts := strings.Split(match.Location, "@")
		clean := func(s string) string {
			return strings.ToLower(strings.ReplaceAll(s, " ", ""))
		}
		if params.Country != "" && (len(locParts) < 1 || clean(locParts[0]) != clean(params.Country)) {
			continue
		}
		if params.City != "" && (len(locParts) < 2 || clean(locParts[1]) != clean(params.City)) {
			continue
		}

		fullname := match.FirstName + " " + match.LastName
		if params.Input != "" {
			target := strings.ToLower(fullname)
			if !strings.HasPrefix(strings.ToLower(target), strings.ToLower(params.Input)) {
				continue
			}
		}

		profs = append(profs, model.FoundProfile{
			IDUser:   int(match.ProfileId),
			FirstImg: match.Photos[0],
			Fullname: fullname,
			Age:      age,
		})
	}

	return profs, nil
}
