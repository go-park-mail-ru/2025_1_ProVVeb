package utils

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/config"
)

type Profile struct {
	Profile config.Profile
}

var (
	ErrInvalidMonth     = errors.New("invalid mounth")
	ErrInvalidDay       = errors.New("invalid day")
	ErrInvalidAge       = errors.New("age range must be between 18 and 100, and from must be less than or equal to to")
	ErrInvalidFirstName = errors.New("first name is required")
	ErrInvalidLastName  = errors.New("last name is required")
	ErrInvalidLocation  = errors.New("location is required")
	ErrInvalidHeight    = errors.New("height must be greater than 0")
	ErrInvalidInterests = errors.New("at least one interest is required")
)

func CompareProfiles(a, b config.Profile) bool {
	return reflect.DeepEqual(a, b)
}

func ValidateBirthday(birthday struct {
	Year  int
	Month int
	Day   int
}) error {
	if birthday.Month < 1 || birthday.Month > 12 {
		return ErrInvalidMonth
	}
	if birthday.Day < 1 || birthday.Day > 31 {
		return ErrInvalidDay
	}

	_, err := time.Parse("2006-01-02", fmt.Sprintf("%d-%02d-%02d", birthday.Year, birthday.Month, birthday.Day))
	if err != nil {
		return err
	}

	return nil
}

func ValidateAge(age struct {
	From int
	To   int
}) error {
	if age.From < 18 || age.To > 100 || age.From > age.To {
		return ErrInvalidAge
	}

	return nil
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

	return nil
}
