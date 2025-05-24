package usecase

import (
	"context"
	"math/rand"
	"time"

	profiles "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/model"
	"github.com/icrowley/fake"
	"github.com/sirupsen/logrus"
)

func (pss *ProfileServiceServer) StoreProfile(
	ctx context.Context,
	req *profiles.StoreProfileRequest,
) (*profiles.StoreProfileResponse, error) {
	pss.Logger.WithFields(&logrus.Fields{"profile": req.Profile}).Info("StoreProfile")
	var fname, lname string

	var likedBy []int
	for _, like := range req.Profile.LikedBy {
		likedBy = append(likedBy, int(like))
	}

	var prefs []model.Preference
	for _, pref := range req.Profile.Preferences {
		prefs = append(prefs, model.Preference{
			Description: pref.Description,
			Value:       pref.Value,
		})
	}

	var params []model.Preference
	for _, pref := range req.Profile.Parametres {
		params = append(params, model.Preference{
			Description: pref.Description,
			Value:       pref.Value,
		})
	}

	var sentProfile model.Profile = model.Profile{
		FirstName:   req.Profile.FirstName,
		LastName:    req.Profile.LastName,
		IsMale:      req.Profile.IsMale,
		Birthday:    req.Profile.Birthday.AsTime(),
		Goal:        int(req.Profile.Goal),
		Height:      int(req.Profile.Height),
		Description: req.Profile.Description,
		Location:    req.Profile.Location,
		Interests:   req.Profile.Interests,
		Photos:      req.Profile.Photos,
		LikedBy:     likedBy,
		Preferences: prefs,
		Parameters:  params,
	}

	if sentProfile.FirstName != "" {
		fname = sentProfile.FirstName
	} else {
		if sentProfile.IsMale {
			fname = fake.MaleFirstName()
		} else {
			fname = fake.FemaleFirstName()
		}
	}

	if sentProfile.LastName != "" {
		lname = sentProfile.LastName
	} else {
		if sentProfile.IsMale {
			lname = fake.MaleLastName()
		} else {
			lname = fake.FemaleLastName()
		}
	}

	var birthday time.Time
	if sentProfile.Birthday.IsZero() {
		birthday = time.Now().AddDate(
			-(rand.Intn(27) + 18),
			-rand.Intn(12),
			-rand.Intn(30),
		)
	} else {
		birthday = sentProfile.Birthday
	}

	height := sentProfile.Height
	if height == 0 {
		height = rand.Intn(50) + 150
	}

	goal := sentProfile.Goal
	if goal == 0 {
		goal = 1
	}

	description := sentProfile.Description
	if description == "" {
		description = fake.SentencesN(2)
	}

	location := sentProfile.Location
	if location == "" {
		location = fake.Country() + "@" + fake.City() + "@" + fake.State()
	}

	interests := sentProfile.Interests
	if len(interests) == 0 {
		for range 5 {
			interests = append(interests, fake.Word())
		}
	}

	var photos []string
	var defaultFileName = "/" + fake.CharactersN(15) + ".png"
	if len(sentProfile.Photos) == 0 {
		photos = make([]string, 0, 6)
		photos = append(photos, defaultFileName)
	} else {
		photos = sentProfile.Photos
	}

	profile := model.Profile{
		FirstName:   fname,
		LastName:    lname,
		IsMale:      sentProfile.IsMale,
		Birthday:    birthday,
		Height:      height,
		Goal:        goal,
		Description: description,
		Location:    location,
		Interests:   interests,
		Photos:      photos,
		Preferences: sentProfile.Preferences,
		Parameters:  sentProfile.Parameters,
		LikedBy:     sentProfile.LikedBy,
	}

	pss.Logger.Info("Profile data generated")

	if len(sentProfile.Photos) == 0 {
		imgBytes, err := pss.StaticRepo.GenerateImage("image/png", sentProfile.IsMale)
		if err != nil {
			pss.Logger.Error("cannot generate image", err)
			return &profiles.StoreProfileResponse{}, err
		}

		err = pss.StaticRepo.UploadImage(imgBytes, defaultFileName, "image/png")
		if err != nil {
			pss.Logger.Error("cannot upload image", err)
			return &profiles.StoreProfileResponse{}, err
		}
	}

	profileId, err := pss.ProfilesRepo.StoreProfile(profile)
	if err != nil {
		pss.Logger.Error("cannot store profile", err)
		return &profiles.StoreProfileResponse{}, err
	}

	err = pss.ProfilesRepo.StorePhotos(profileId, photos)
	if err != nil {
		pss.Logger.Error("cannot store photos", err)
		return &profiles.StoreProfileResponse{}, err
	}

	err = pss.ProfilesRepo.StoreInterests(profileId, interests)
	if err != nil {
		pss.Logger.Error("cannot store interests", err)
		return &profiles.StoreProfileResponse{}, err
	}
	pss.Logger.WithFields(&logrus.Fields{"profileId": profileId}).Info("Profile saved")

	return &profiles.StoreProfileResponse{
		ProfileId: int32(profileId),
	}, nil
}
