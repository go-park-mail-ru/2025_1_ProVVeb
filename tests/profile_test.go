package tests

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/utils"
)

func TestValidateBirthday(t *testing.T) {
	tests := []struct {
		birthday struct {
			Year  int
			Month int
			Day   int
		}
		expectedError error
	}{
		{birthday: struct{ Year, Month, Day int }{2005, 1, 1}, expectedError: nil},
		{birthday: struct{ Year, Month, Day int }{1995, 5, 31}, expectedError: nil},
		{birthday: struct{ Year, Month, Day int }{2005, 9, 32}, expectedError: utils.ErrInvalidDay},
		{birthday: struct{ Year, Month, Day int }{2003, -1, 2}, expectedError: utils.ErrInvalidMonth},
		{birthday: struct{ Year, Month, Day int }{1234, 11, 6}, expectedError: nil},
		{birthday: struct{ Year, Month, Day int }{1987, 32, 2}, expectedError: utils.ErrInvalidMonth},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d-%d-%d", tt.birthday.Year, tt.birthday.Month, tt.birthday.Day), func(t *testing.T) {
			err := utils.ValidateBirthday(tt.birthday)
			if err != nil && !errors.Is(err, tt.expectedError) {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}
			if err == nil && tt.expectedError != nil {
				t.Errorf("expected error %v, but got none", tt.expectedError)
			}
		})
	}
}

func TestValidateAge(t *testing.T) {
	tests := []struct {
		age           struct{ From, To int }
		expectedError error
	}{
		{age: struct{ From, To int }{18, 30}, expectedError: nil},
		{age: struct{ From, To int }{23, 56}, expectedError: nil},
		{age: struct{ From, To int }{98, 102}, expectedError: utils.ErrInvalidAge},
		{age: struct{ From, To int }{3, 4}, expectedError: utils.ErrInvalidAge},
		{age: struct{ From, To int }{78, 34}, expectedError: utils.ErrInvalidAge},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d-%d", tt.age.From, tt.age.To), func(t *testing.T) {
			err := utils.ValidateAge(tt.age)
			if err != nil && !errors.Is(tt.expectedError, err) {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}
			if err == nil && tt.expectedError != nil {
				t.Errorf("expected error %v, but got none", tt.expectedError)
			}
		})
	}
}

// func TestValidateProfile(t *testing.T) {
// 	tests := []struct {
// 		profile       utils.Profile
// 		expectedError error
// 	}{
// 		{profile: utils.Profile{Profile: config.Profile{
// 			ProfileId: 1,
// 			FirstName: "John",
// 			LastName:  "Doe",
// 			Height:    180,
// 			Birthday: struct {
// 				Year  int `yaml:"year" json:"year"`
// 				Month int `yaml:"month" json:"month"`
// 				Day   int `yaml:"day" json:"day"`
// 			}{
// 				Year:  2000,
// 				Month: 5,
// 				Day:   20,
// 			},
// 			Location:  "USA",
// 			Interests: []string{"Reading", "Traveling"},
// 			Preferences: struct {
// 				PreferencesId int      `yaml:"preferencesId" json:"preferencesId"`
// 				Interests     []string `yaml:"interests" json:"interests"`
// 				Location      string   `yaml:"location" json:"location"`
// 				Age           struct {
// 					From int `yaml:"from" json:"from"`
// 					To   int `yaml:"to" json:"to"`
// 				}
// 			}{
// 				PreferencesId: 1,
// 				Interests:     []string{"Sports", "Technology"},
// 				Location:      "USA",
// 				Age: struct {
// 					From int `yaml:"from" json:"from"`
// 					To   int `yaml:"to" json:"to"`
// 				}{From: 18, To: 35},
// 			},
// 		}}, expectedError: nil},

// 		{profile: utils.Profile{Profile: config.Profile{
// 			ProfileId: 2,
// 			FirstName: "",
// 			LastName:  "Smith",
// 			Height:    170,
// 			Birthday: struct {
// 				Year  int `yaml:"year" json:"year"`
// 				Month int `yaml:"month" json:"month"`
// 				Day   int `yaml:"day" json:"day"`
// 			}{
// 				Year:  1990,
// 				Month: 6,
// 				Day:   15,
// 			},
// 			Location:  "UK",
// 			Interests: []string{"Music", "Traveling"},
// 		}}, expectedError: utils.ErrInvalidFirstName},

// 		{profile: utils.Profile{Profile: config.Profile{
// 			ProfileId: 3,
// 			FirstName: "Jane",
// 			LastName:  "Doe",
// 			Height:    165,
// 			Birthday: struct {
// 				Year  int `yaml:"year" json:"year"`
// 				Month int `yaml:"month" json:"month"`
// 				Day   int `yaml:"day" json:"day"`
// 			}{
// 				Year:  1985,
// 				Month: 7,
// 				Day:   25,
// 			},
// 			Location:  "Germany",
// 			Interests: []string{"Photography", "Cycling"},
// 			Preferences: struct {
// 				PreferencesId int      `yaml:"preferencesId" json:"preferencesId"`
// 				Interests     []string `yaml:"interests" json:"interests"`
// 				Location      string   `yaml:"location" json:"location"`
// 				Age           struct {
// 					From int `yaml:"from" json:"from"`
// 					To   int `yaml:"to" json:"to"`
// 				}
// 			}{
// 				PreferencesId: 2,
// 				Interests:     []string{"Music", "Art"},
// 				Location:      "Germany",
// 				Age: struct {
// 					From int `yaml:"from" json:"from"`
// 					To   int `yaml:"to" json:"to"`
// 				}{From: 18, To: 30},
// 			},
// 		}}, expectedError: nil},

// 		{profile: utils.Profile{Profile: config.Profile{
// 			ProfileId: 4,
// 			FirstName: "Alice",
// 			LastName:  "Johnson",
// 			Height:    175,
// 			Birthday: struct {
// 				Year  int `yaml:"year" json:"year"`
// 				Month int `yaml:"month" json:"month"`
// 				Day   int `yaml:"day" json:"day"`
// 			}{
// 				Year:  1992,
// 				Month: 3,
// 				Day:   10,
// 			},
// 			Location:  "Canada",
// 			Interests: []string{"Dancing", "Art"},
// 			Preferences: struct {
// 				PreferencesId int      `yaml:"preferencesId" json:"preferencesId"`
// 				Interests     []string `yaml:"interests" json:"interests"`
// 				Location      string   `yaml:"location" json:"location"`
// 				Age           struct {
// 					From int `yaml:"from" json:"from"`
// 					To   int `yaml:"to" json:"to"`
// 				}
// 			}{
// 				PreferencesId: 3,
// 				Interests:     []string{"Sports", "Music"},
// 				Location:      "Canada",
// 				Age: struct {
// 					From int `yaml:"from" json:"from"`
// 					To   int `yaml:"to" json:"to"`
// 				}{From: 35, To: 30},
// 			},
// 		}}, expectedError: utils.ErrInvalidAge},

// 		{profile: utils.Profile{Profile: config.Profile{
// 			ProfileId: 5,
// 			FirstName: "Bob",
// 			LastName:  "Brown",
// 			Height:    160,
// 			Birthday: struct {
// 				Year  int `yaml:"year" json:"year"`
// 				Month int `yaml:"month" json:"month"`
// 				Day   int `yaml:"day" json:"day"`
// 			}{
// 				Year:  1993,
// 				Month: 8,
// 				Day:   17,
// 			},
// 			Location:  "Australia",
// 			Interests: []string{},
// 		}}, expectedError: utils.ErrInvalidInterests},
// 	}

// 	for _, tt := range tests {
// 		t.Run(fmt.Sprintf("ProfileId %d", tt.profile.Profile.ProfileId), func(t *testing.T) {
// 			err := utils.ValidateProfile(tt.profile)
// 			if err != nil && !errors.Is(tt.expectedError, err) {
// 				t.Errorf("expected error %v, got %v", tt.expectedError, err)
// 			}
// 			if err == nil && tt.expectedError != nil {
// 				t.Errorf("expected error %v, but got none", tt.expectedError)
// 			}
// 		})
// 	}
// }
