package repository

import (
	"database/sql"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetProfileById(t *testing.T) {
	mockDB := new(MockDB)
	rows := &MockRows{
		data: [][]interface{}{
			{
				1, "Alice", "Smith", true, 170,
				sql.NullTime{Time: time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true},
				"Description", sql.NullInt64{Int64: 1, Valid: true},
				sql.NullString{String: "USA", Valid: true},
				sql.NullString{String: "NYC", Valid: true},
				sql.NullString{String: "Brooklyn", Valid: true},
				sql.NullInt64{Int64: 2, Valid: true},
				sql.NullString{String: "/images/a.jpg", Valid: true},
				sql.NullString{String: "Music", Valid: true},
				sql.NullString{String: "Smoke", Valid: true},
				sql.NullString{String: "Never", Valid: true},
				sql.NullString{String: "Height", Valid: true},
				sql.NullString{String: "170", Valid: true},
				sql.NullBool{Bool: true, Valid: true},
				sql.NullInt64{Int64: 5, Valid: true},
			},
		},
	}

	mockDB.On("Query", mock.Anything, repository.GetProfileByIdQuery, []interface{}{1}).Return(rows, nil)

	repo := &repository.ProfileRepo{DB: mockDB}

	profile, err := repo.GetProfileById(1)

	assert.NoError(t, err)
	assert.Equal(t, 1, profile.ProfileId)
	assert.Equal(t, "Alice", profile.FirstName)
	assert.Equal(t, "USA@NYC@Brooklyn", profile.Location)
	assert.Contains(t, profile.LikedBy, 2)
	assert.Contains(t, profile.Interests, "Music")
	assert.Equal(t, true, profile.Premium.Status)
	assert.Equal(t, 5, profile.Premium.Border)
}

func TestGetRecommendations(t *testing.T) {
	mockDB := new(MockDB)

	rows := &MockRows{
		data: [][]interface{}{
			{
				1,
				"Alice",
				"Smith",
				true,
				170,
				sql.NullTime{Time: time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true},
				"Test Description",
				sql.NullInt64{Int64: 1, Valid: true},
				sql.NullString{String: "USA", Valid: true},
				sql.NullString{String: "NYC", Valid: true},
				sql.NullString{String: "Brooklyn", Valid: true},
				sql.NullInt64{Int64: 2, Valid: true},
				sql.NullString{String: "/img/avatar.jpg", Valid: true},
				sql.NullString{String: "Music", Valid: true},
				sql.NullString{String: "Smoking", Valid: true},
				sql.NullString{String: "Never", Valid: true},
				sql.NullString{String: "Height", Valid: true},
				sql.NullString{String: "170", Valid: true},
				sql.NullBool{Bool: true, Valid: true},
				sql.NullInt64{Int64: 3, Valid: true},
			},
		},
	}

	mockDB.On("Query", mock.Anything, repository.GetRecommendationsQuery, []interface{}{1}).Return(rows, nil)

	repo := &repository.ProfileRepo{DB: mockDB}

	profile, err := repo.GetRecomendations(1)

	assert.NoError(t, err)
	assert.Equal(t, 1, profile.ProfileId)
	assert.Equal(t, "Alice", profile.FirstName)
	assert.Equal(t, "Smith", profile.LastName)
	assert.Equal(t, true, profile.IsMale)
	assert.Equal(t, 170, profile.Height)
	assert.Equal(t, time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC), profile.Birthday)
	assert.Equal(t, "Test Description", profile.Description)
	assert.Equal(t, 1, profile.Goal)
	assert.Equal(t, "USA@NYC@Brooklyn", profile.Location)
	assert.Contains(t, profile.LikedBy, 2)
	assert.Contains(t, profile.Interests, "Music")
	assert.Contains(t, profile.Preferences, model.Preference{
		Description: "Smoking",
		Value:       "Never",
	})
	assert.Contains(t, profile.Parameters, model.Preference{
		Description: "Height",
		Value:       "170",
	})
	assert.Contains(t, profile.Photos, "/img/avatar.jpg")
	assert.True(t, profile.Premium.Status)
	assert.Equal(t, 3, profile.Premium.Border)
}

func TestGetProfileStats(t *testing.T) {
	mockDB := new(MockDB)

	expectedProfileID := 1
	expectedStats := []interface{}{
		5,
		10,
		3,
		1,
		2,
		50,
		7,
	}

	row := &mockRow{
		values: expectedStats,
	}

	mockDB.On("QueryRow", mock.Anything, repository.GetStaticticsQuery, []interface{}{expectedProfileID}).Return(row)

	repo := &repository.ProfileRepo{DB: mockDB}

	stats, err := repo.GetProfileStats(expectedProfileID)

	assert.NoError(t, err)
	assert.Equal(t, 5, stats.LikesGiven)
	assert.Equal(t, 10, stats.LikesReceived)
	assert.Equal(t, 3, stats.Matches)
	assert.Equal(t, 1, stats.ComplaintsMade)
	assert.Equal(t, 2, stats.ComplaintsReceived)
	assert.Equal(t, 50, stats.MessagesSent)
	assert.Equal(t, 7, stats.ChatCount)
}

func TestSearchProfiles(t *testing.T) {
	mockDB := new(MockDB)

	searchParams := model.SearchProfileRequest{
		IsMale:    "Any",
		AgeMin:    18,
		AgeMax:    30,
		HeightMin: 160,
		HeightMax: 190,
		Goal:      1,
		Country:   "USA",
		City:      "NYC",
		Input:     "Ali",
		Preferences: []model.Preference{
			{
				Description: "Height",
				Value:       "170",
			},
		},
	}

	rows := &MockRows{
		data: [][]interface{}{
			{
				1,
				"/img/ava1.jpg",
				"Alice Smith",
				25,
				1,
			},
			{
				2,
				"/img/ava2.jpg",
				"Bob Johnson",
				28,
				1,
			},
		},
	}

	mockDB.On("Query", mock.Anything, repository.SearchProfilesQuery, []interface{}{
		1,
		searchParams.IsMale,
		searchParams.AgeMin,
		searchParams.AgeMax,
		searchParams.HeightMin,
		searchParams.HeightMax,
		searchParams.Goal,
		searchParams.Country,
		searchParams.City,
		searchParams.Preferences,
		searchParams.Input,
	}).Return(rows, nil)

	repo := &repository.ProfileRepo{DB: mockDB}

	results, err := repo.SearchProfiles(1, searchParams)

	assert.NoError(t, err)
	assert.Len(t, results, 2)

	assert.Equal(t, 1, results[0].IDUser)
	assert.Equal(t, "/img/ava1.jpg", results[0].FirstImg)
	assert.Equal(t, "Alice Smith", results[0].Fullname)
	assert.Equal(t, 25, results[0].Age)
	assert.Equal(t, 1, results[0].Goal)

	assert.Equal(t, 2, results[1].IDUser)
	assert.Equal(t, "Bob Johnson", results[1].Fullname)
}
