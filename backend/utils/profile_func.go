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

	err := ValidateBirthday(struct {
		Year  int
		Month int
		Day   int
	}(p.Profile.Birthday))
	if err != nil {
		return err
	}

	if profile.Height <= 0 {
		return ErrInvalidHeight
	}

	if len(profile.Interests) == 0 {
		return ErrInvalidInterests
	}

	err = ValidateAge(struct {
		From int
		To   int
	}(p.Profile.Preferences.Age))
	if err != nil {
		return err
	}

	return nil
}

func InitProfileMap() map[int]config.Profile {
	profiles := make(map[int]config.Profile)

	profile1 := config.Profile{
		ProfileId: 1,
		FirstName: "Лиза",
		LastName:  "Тимофеева",
		Height:    180,
		Birthday: struct {
			Year  int `yaml:"year" json:"year"`
			Month int `yaml:"month" json:"month"`
			Day   int `yaml:"day" json:"day"`
		}{
			Year:  1990,
			Month: 5,
			Day:   15,
		},
		Avatar:      "avatars/liza.png",
		Card:        "cards/liza.png",
		Description: "Специалист по IT",
		Location:    "New York",
		Interests:   []string{"Technology", "Reading", "Traveling"},
		LikedBy:     []int{2, 3, 4},
		Preferences: struct {
			PreferencesId int      `yaml:"preferencesId" json:"preferencesId"`
			Interests     []string `yaml:"interests" json:"interests"`
			Location      string   `yaml:"location" json:"location"`
			Age           struct {
				From int `yaml:"from" json:"from"`
				To   int `yaml:"to" json:"to"`
			}
		}{
			PreferencesId: 1,
			Interests:     []string{"Music", "Movies", "Sports"},
			Location:      "New York",
			Age: struct {
				From int `yaml:"from" json:"from"`
				To   int `yaml:"to" json:"to"`
			}{
				From: 18,
				To:   35,
			},
		},
	}

	err := ValidateProfile(Profile{Profile: profile1})
	if err != nil {
		fmt.Println("Error validating profile 1:", err)
	} else {
		profiles[profile1.ProfileId] = profile1
	}

	profile2 := config.Profile{
		ProfileId: 2,
		FirstName: "Катя",
		LastName:  "Кострицкая",
		Height:    165,
		Birthday: struct {
			Year  int `yaml:"year" json:"year"`
			Month int `yaml:"month" json:"month"`
			Day   int `yaml:"day" json:"day"`
		}{
			Year:  1995,
			Month: 8,
			Day:   22,
		},
		Avatar:      "avatars/katya.png",
		Card:        "cards/katya.png",
		Description: "Любитель хорошего кода",
		Location:    "California",
		Interests:   []string{"Hiking", "Photography", "Art"},
		LikedBy:     []int{1, 3, 5},
		Preferences: struct {
			PreferencesId int      `yaml:"preferencesId" json:"preferencesId"`
			Interests     []string `yaml:"interests" json:"interests"`
			Location      string   `yaml:"location" json:"location"`
			Age           struct {
				From int `yaml:"from" json:"from"`
				To   int `yaml:"to" json:"to"`
			}
		}{
			PreferencesId: 2,
			Interests:     []string{"Art", "Nature", "Traveling"},
			Location:      "California",
			Age: struct {
				From int `yaml:"from" json:"from"`
				To   int `yaml:"to" json:"to"`
			}{
				From: 20,
				To:   40,
			},
		},
	}

	err = ValidateProfile(Profile{Profile: profile2})
	if err != nil {
		fmt.Println("Error validating profile 2:", err)
	} else {
		profiles[profile2.ProfileId] = profile2
	}

	profile3 := config.Profile{
		ProfileId: 3,
		FirstName: "Ева",
		LastName:  "Ильченко",
		Height:    170,
		Birthday: struct {
			Year  int `yaml:"year" json:"year"`
			Month int `yaml:"month" json:"month"`
			Day   int `yaml:"day" json:"day"`
		}{
			Year:  1992,
			Month: 2,
			Day:   10,
		},
		Avatar:      "avarars/eva.png",
		Card:        "cards/eva.png",
		Description: "Студентка ИУ7",
		Location:    "Los Angeles",
		Interests:   []string{"Cooking", "Traveling", "Fitness"},
		LikedBy:     []int{1, 2, 4},
		Preferences: struct {
			PreferencesId int      `yaml:"preferencesId" json:"preferencesId"`
			Interests     []string `yaml:"interests" json:"interests"`
			Location      string   `yaml:"location" json:"location"`
			Age           struct {
				From int `yaml:"from" json:"from"`
				To   int `yaml:"to" json:"to"`
			}
		}{
			PreferencesId: 3,
			Interests:     []string{"Food", "Traveling", "Health"},
			Location:      "Los Angeles",
			Age: struct {
				From int `yaml:"from" json:"from"`
				To   int `yaml:"to" json:"to"`
			}{
				From: 18,
				To:   45,
			},
		},
	}

	err = ValidateProfile(Profile{Profile: profile3})
	if err != nil {
		fmt.Println("Error validating profile 3:", err)
	} else {
		profiles[profile3.ProfileId] = profile3
	}

	return profiles
}
