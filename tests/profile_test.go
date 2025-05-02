package tests

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/utils"
	"github.com/stretchr/testify/require"
)

func TestValidateProfile(t *testing.T) {
	tests := []struct {
		profile       utils.Profile
		expectedError error
	}{
		{profile: utils.Profile{Profile: model.Profile{
			ProfileId: 1,
			FirstName: "John",
			LastName:  "Doe",
			Height:    180,
			Birthday:  time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
			Location:  "USA",
			Interests: []string{"Reading", "Traveling"},
		}}, expectedError: nil},

		{profile: utils.Profile{Profile: model.Profile{
			ProfileId: 2,
			FirstName: "",
			LastName:  "Smith",
			Height:    170,
			Birthday:  time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
			Location:  "UK",
			Interests: []string{"Music", "Traveling"},
		}}, expectedError: utils.ErrInvalidFirstName},

		{profile: utils.Profile{Profile: model.Profile{
			ProfileId: 3,
			FirstName: "Jane",
			LastName:  "Doe",
			Height:    165,
			Birthday:  time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
			Location:  "Germany",
			Interests: []string{"Photography", "Cycling"},
		}}, expectedError: nil},

		{profile: utils.Profile{Profile: model.Profile{
			ProfileId: 4,
			FirstName: "Bob",
			LastName:  "Brown",
			Height:    160,
			Birthday:  time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
			Location:  "Australia",
			Interests: []string{},
		}}, expectedError: utils.ErrInvalidInterests},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("ProfileId %d", tt.profile.Profile.ProfileId), func(t *testing.T) {
			err := utils.ValidateProfile(tt.profile)
			if err != nil && !errors.Is(tt.expectedError, err) {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}
			if err == nil && tt.expectedError != nil {
				t.Errorf("expected error %v, but got none", tt.expectedError)
			}
		})
	}
}

func TestCreateUser(t *testing.T) {
	tests := []struct {
		id           int
		login        string
		password     string
		expectedUser model.User
		expectedErr  string
	}{
		{
			id:          -5,
			login:       "user",
			password:    "StrongPass123!",
			expectedErr: "invalid id",
		},
		{
			id:          2,
			login:       "x",
			password:    "StrongPass123!",
			expectedErr: "login",
		},
		{
			id:          3,
			login:       "validUser",
			password:    "123",
			expectedErr: "password",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("id=%d_login=%s", tt.id, tt.login), func(t *testing.T) {
			user, err := utils.CreateUser(tt.id, tt.login, tt.password)

			if tt.expectedErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedUser.UserId, user.User.UserId)
				require.Equal(t, tt.expectedUser.Login, user.User.Login)
				require.NotEqual(t, tt.password, user.User.Password, "password should be encrypted")
				require.NotEmpty(t, user.User.Password, "encrypted password should not be empty")
			}
		})
	}
}
