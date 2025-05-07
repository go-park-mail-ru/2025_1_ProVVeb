package tests

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func UserRepoTestGetProfileById(t *testing.T) {
	tests := []struct {
		name        string
		profileID   int
		mockRows    func(mock sqlmock.Sqlmock)
		expected    model.Profile
		expectedErr bool
	}{
		{
			name:      "valid profile with all fields",
			profileID: 1,
			mockRows: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(repository.GetProfileByIdQuery)).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{
						"profile_id", "firstname", "lastname", "is_male", "height",
						"birthday", "description", "country", "city", "district",
						"liked_by_profile_id", "avatar", "interest", "preference_description", "preference_value",
					}).AddRow(
						1, "Иван", "Иванов", true, 180,
						time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC), "Описание", "Россия", "Москва", "ЦАО",
						sql.NullInt64{Int64: 2, Valid: true},
						sql.NullString{String: "avatar.jpg", Valid: true},
						sql.NullString{String: "Музыка", Valid: true},
						sql.NullString{String: "Рост", Valid: true},
						sql.NullString{String: "180+", Valid: true},
					))
			},
			expected: model.Profile{
				ProfileId:   1,
				FirstName:   "Иван",
				LastName:    "Иванов",
				IsMale:      true,
				Height:      180,
				Birthday:    time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
				Description: "Описание",
				Location:    "Россия@Москва@ЦАО",
				LikedBy:     []int{2},
				Photos:      []string{"avatar.jpg"},
				Interests:   []string{"Музыка"},
				Preferences: []model.Preference{
					{Description: "Рост", Value: "180+"},
				},
			},
			expectedErr: false,
		},
		{
			name:      "empty profile (no data)",
			profileID: 42,
			mockRows: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(repository.GetProfileByIdQuery)).
					WithArgs(42).
					WillReturnRows(sqlmock.NewRows([]string{
						"profile_id", "firstname", "lastname", "is_male", "height",
						"birthday", "description", "country", "city", "district",
						"liked_by_profile_id", "avatar", "interest", "preference_description", "preference_value",
					}))
			},
			expected:    model.Profile{},
			expectedErr: false,
		},
		{
			name:      "query error",
			profileID: 99,
			mockRows: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(repository.GetProfileByIdQuery)).
					WithArgs(99).
					WillReturnError(sql.ErrConnDone)
			},
			expected:    model.Profile{},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			tt.mockRows(mock)

			repo := &repository.UserRepo{DB: db}
			profile, err := repo.GetProfileById(tt.profileID)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.ProfileId, profile.ProfileId)
				assert.Equal(t, tt.expected.FirstName, profile.FirstName)
				assert.Equal(t, tt.expected.LastName, profile.LastName)
				assert.Equal(t, tt.expected.IsMale, profile.IsMale)
				assert.Equal(t, tt.expected.Height, profile.Height)
				assert.Equal(t, tt.expected.Birthday, profile.Birthday)
				assert.Equal(t, tt.expected.Description, profile.Description)
				assert.Equal(t, tt.expected.Location, profile.Location)
				assert.ElementsMatch(t, tt.expected.LikedBy, profile.LikedBy)
				assert.ElementsMatch(t, tt.expected.Photos, profile.Photos)
				assert.ElementsMatch(t, tt.expected.Interests, profile.Interests)
				assert.ElementsMatch(t, tt.expected.Preferences, profile.Preferences)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func UserRepoTestSQL_StoreInterests(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &repository.UserRepo{DB: db}

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT interest_id FROM interests WHERE description = \$1`).
		WithArgs("Музыка").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery(`INSERT INTO interests \(description\) VALUES \(\$1\) RETURNING interest_id`).
		WithArgs("Музыка").
		WillReturnRows(sqlmock.NewRows([]string{"interest_id"}).AddRow(3))

	mock.ExpectExec(`INSERT INTO profile_interests \(profile_id, interest_id\) VALUES \(\$1, \$2\)`).
		WithArgs(1, 3).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = repo.StoreInterests(1, []string{"Музыка"})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestComplaintRepo_CreateComplaint(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &repository.ComplaintRepo{DB: db}

	const (
		complaintBy   = 1
		complaintOn   = 2
		complaintType = "Спам"
		text          = "Он мне пишет рекламу"
	)
	var complaintTypeID = 42

	mock.ExpectQuery("SELECT comp_type FROM complaint_types").
		WithArgs(complaintType).
		WillReturnRows(sqlmock.NewRows([]string{"comp_type"}).AddRow(complaintTypeID))

	mock.ExpectExec("INSERT INTO complaints").
		WithArgs(complaintBy, complaintOn, complaintTypeID, text, 1, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateComplaint(complaintBy, complaintOn, complaintType, text)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestComplaintRepo_CreateComplaint_TypeNotExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &repository.ComplaintRepo{DB: db}

	const (
		complaintBy   = 3
		complaintType = "Нарушение правил"
		text          = "Неприемлемое поведение"
	)
	var insertedTypeID = 99

	mock.ExpectQuery("SELECT comp_type FROM complaint_types").
		WithArgs(complaintType).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery("INSERT INTO complaint_types").
		WithArgs(complaintType).
		WillReturnRows(sqlmock.NewRows([]string{"comp_type"}).AddRow(insertedTypeID))

	mock.ExpectExec("INSERT INTO complaints").
		WithArgs(complaintBy, complaintBy, insertedTypeID, text, 1, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateComplaint(complaintBy, 0, complaintType, text)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
