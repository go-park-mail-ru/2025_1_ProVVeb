package utils

import (
	"errors"
	"reflect"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
)

type Profile struct {
	Profile model.Profile
}

var (
	ErrInvalidMonth        = errors.New("invalid mounth")
	ErrInvalidDay          = errors.New("invalid day")
	ErrInvalidAge          = errors.New("age range must be between 18 and 100, and from must be less than or equal to to")
	ErrInvalidFirstName    = errors.New("first name is required")
	ErrInvalidLastName     = errors.New("last name is required")
	ErrInvalidLocation     = errors.New("location is required")
	ErrInvalidHeight       = errors.New("height must be greater than 0")
	ErrInvalidInterests    = errors.New("at least one interest is required")
	ErrInvalidIPreferences = errors.New("at least one preference is required")
)

func CompareProfiles(a, b model.Profile) bool {
	return reflect.DeepEqual(a, b)
}

func ValidateProfile(p Profile) error {
	profile := p.Profile
	if profile.FirstName == "" {
		return ErrInvalidFirstName
	}
	if profile.LastName == "" {
		return ErrInvalidLastName
	}
	if profile.Location == "" {
		return ErrInvalidLocation
	}

	if len(profile.Interests) == 0 {
		return ErrInvalidInterests
	}

	if len(profile.Preferences) == 0 {
		return ErrInvalidIPreferences
	}

	return nil
}

func CalculateAge(birthday time.Time) int {
	now := time.Now()
	age := now.Year() - birthday.Year()
	if now.YearDay() < birthday.YearDay() {
		age--
	}
	return age
}
