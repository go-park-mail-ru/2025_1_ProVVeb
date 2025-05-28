package repository

import (
	"database/sql"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetProfileById(t *testing.T) {
	mockDB := new(MockDB)

	// Пример строки результата
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
