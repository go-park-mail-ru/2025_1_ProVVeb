package tests

import (
	"database/sql"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/repository"
	"github.com/stretchr/testify/assert"
	mockk "github.com/stretchr/testify/mock"
)

func TestGetProfileById(t *testing.T) {
	tests := []struct {
		name        string
		profileID   int
		mockRows    func(mock *MockDB)
		expected    model.Profile
		expectedErr bool
	}{
		{
			name:      "valid profile with all fields",
			profileID: 1,
			mockRows: func(mock *MockDB) {
				rows := &MockRows{
					data: [][]interface{}{
						{
							1,        // profile_id
							"Иван",   // firstname
							"Иванов", // lastname
							true,     // is_male
							180,      // height
							sql.NullTime{ // birthday
								Time:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
								Valid: true,
							},
							sql.NullString{String: "Описание", Valid: true},   // description
							sql.NullInt64{Int64: 1, Valid: true},              // goal
							sql.NullString{String: "Россия", Valid: true},     // country
							sql.NullString{String: "Москва", Valid: true},     // city
							sql.NullString{String: "ЦАО", Valid: true},        // district
							sql.NullInt64{Int64: 2, Valid: true},              // liked_by_profile_id
							sql.NullString{String: "avatar.jpg", Valid: true}, // avatar
							sql.NullString{String: "Музыка", Valid: true},     // interest
							sql.NullString{String: "Рост", Valid: true},       // preference_description
							sql.NullString{String: "180+", Valid: true},       // preference_value
							sql.NullString{String: "Параметр", Valid: true},   // parameter_description
							sql.NullString{String: "Значение", Valid: true},   // parameter_value
							sql.NullBool{Bool: true, Valid: true},             // premium_status
							sql.NullInt64{Int64: 1, Valid: true},              // border
						},
					},
				}
				mock.On("Query", mockk.Anything, repository.GetProfileByIdQuery, []interface{}{1}).Return(rows, nil)
			},
			expected: model.Profile{
				ProfileId:   1,
				FirstName:   "Иван",
				LastName:    "Иванов",
				IsMale:      true,
				Height:      180,
				Birthday:    time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
				Description: "Описание",
				Goal:        1,
				Location:    "Россия@Москва@ЦАО",
				LikedBy:     []int{2},
				Photos:      []string{"avatar.jpg"},
				Interests:   []string{"Музыка"},
				Preferences: []model.Preference{
					{Description: "Рост", Value: "180+"},
				},
				Parameters: []model.Preference{
					{Description: "Параметр", Value: "Значение"},
				},
				Premium: struct {
					Status bool `yaml:"Status" json:"Status"`
					Border int  `yaml:"Border" json:"Border"`
				}{
					Status: true,
					Border: 1,
				},
			},
			expectedErr: false,
		},
		{
			name:      "empty profile (no data)",
			profileID: 42,
			mockRows: func(mock *MockDB) {
				rows := &MockRows{data: [][]interface{}{}}
				mock.On("Query", mockk.Anything, repository.GetProfileByIdQuery, []interface{}{42}).Return(rows, nil)
			},
			expected:    model.Profile{},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDB)
			tt.mockRows(mockDB)
			repo := &repository.ProfileRepo{DB: mockDB}
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
			mockDB.AssertExpectations(t)
		})
	}
}

// func TestSQL_StoreInterests(t *testing.T) {
// 	mockDB := new(MockDB)
// 	repo := &repository.ProfileRepo{DB: mockDB}

//     mockDB.On("Begin", mock.Anything).Return(mockDB, nil)
// 	mockDB.On("Rollback", mock.Anything).Return(nil).Maybe()
// 	mockDB.On("QueryRow", mock.Anything,
// 		`
// SELECT interest_id FROM interests WHERE description = $1
// `,
// 		[]interface{}{"Музыка"}).
// 		Return(&MockRows{}, sql.ErrNoRows)

// 	mockDB.On("QueryRow", mock.Anything,
// 		`
// INSERT INTO interests (description) VALUES ($1) RETURNING interest_id
// `,
// 		[]interface{}{"Музыка"}).
// 		Return(&MockRows{data: [][]interface{}{{3}}}, nil)

// 	mockDB.On("Exec", mock.Anything,
// 		`INSERT INTO profile_interests (profile_id, interest_id) VALUES ($1, $2)`,
// 		[]interface{}{1, 3}).
// 		Return(pgconn.NewCommandTag("INSERT 0 1"), nil)

// 	err := repo.StoreInterests(1, []string{"Музыка"})
// 	require.NoError(t, err)
// 	mockDB.AssertExpectations(t)
// }

const (
	GetInterestIdByDescription = `
SELECT interest_id FROM interests WHERE description = $1
`
	InsertInterestIfNotExists = `
INSERT INTO interests (description)
VALUES ($1)
RETURNING interest_id
`
	InsertProfileInterest = `
INSERT INTO profile_interests (profile_id, interest_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING
`
)

// func TestStoreInterests(t *testing.T) {
// 	t.Run("successful interest storage with new and existing interests", func(t *testing.T) {
// 		// Setup
// 		mockDB := new(MockDB)
// 		repo := &repository.ProfileRepo{DB: mockDB}

// 		profileID := 1
// 		interests := []string{"music", "sports", "reading"}

// 		// Mock expectations
// 		// Begin transaction
// 		mockTx := new(MockDB)
// 		mockDB.On("Begin", context.Background()).Return(mockTx, nil)

// 		// First interest - exists
// 		mockRow1 := NewMockRow([]interface{}{42}) // interest_id = 42
// 		mockTx.On("QueryRow", context.Background(), GetInterestIdByDescription, []interface{}{"music"}).Return(mockRow1)

// 		// Second interest - doesn't exist
// 		mockRowEmpty := &MockRowWithError{err: pgx.ErrNoRows}
// 		mockTx.On("QueryRow", context.Background(), GetInterestIdByDescription, []interface{}{"sports"}).Return(mockRowEmpty)

// 		// Then inserted
// 		mockRowInsert := NewMockRow([]interface{}{73}) // new interest_id = 73
// 		mockTx.On("QueryRow", context.Background(), InsertInterestIfNotExists, []interface{}{"sports"}).Return(mockRowInsert)

// 		// Third interest - exists
// 		mockRow3 := NewMockRow([]interface{}{55}) // interest_id = 55
// 		mockTx.On("QueryRow", context.Background(), GetInterestIdByDescription, []interface{}{"reading"}).Return(mockRow3)

// 		// Profile interest inserts
// 		mockTx.On("Exec", context.Background(), InsertProfileInterest, []interface{}{profileID, 42}).Return(pgconn.NewCommandTag("INSERT 0 1"), nil)
// 		mockTx.On("Exec", context.Background(), InsertProfileInterest, []interface{}{profileID, 73}).Return(pgconn.NewCommandTag("INSERT 0 1"), nil)
// 		mockTx.On("Exec", context.Background(), InsertProfileInterest, []interface{}{profileID, 55}).Return(pgconn.NewCommandTag("INSERT 0 1"), nil)

// 		// Commit
// 		mockTx.On("Commit", context.Background()).Return(nil)
// 		mockTx.On("Rollback", context.Background()).Return(nil)

// 		// Execute
// 		err := repo.StoreInterests(profileID, interests)

// 		// Verify
// 		assert.NoError(t, err)
// 		mockDB.AssertExpectations(t)
// 		mockTx.AssertExpectations(t)
// 	})

// 	t.Run("failed to begin transaction", func(t *testing.T) {
// 		mockDB := new(MockDB)
// 		repo := &repository.ProfileRepo{DB: mockDB}

// 		expectedErr := errors.New("connection error")
// 		mockDB.On("Begin", context.Background()).Return(nil, expectedErr)

// 		err := repo.StoreInterests(1, []string{"music"})
// 		assert.EqualError(t, err, expectedErr.Error())
// 		mockDB.AssertExpectations(t)
// 	})

// 	t.Run("failed to insert new interest", func(t *testing.T) {
// 		mockDB := new(MockDB)
// 		repo := &repository.ProfileRepo{DB: mockDB}

// 		profileID := 1
// 		interests := []string{"unknown"}

// 		mockTx := new(MockDB)
// 		mockDB.On("Begin", context.Background()).Return(mockTx, nil)

// 		emptyRows := &MockRows{}
// 		mockTx.On("QueryRow", context.Background(), GetInterestIdByDescription, []interface{}{"unknown"}).Return(emptyRows)

// 		expectedErr := errors.New("insert failed")
// 		mockTx.On("QueryRow", context.Background(), InsertInterestIfNotExists, []interface{}{"unknown"}).Return(nil, expectedErr)

// 		mockTx.On("Rollback", context.Background()).Return(nil)

// 		err := repo.StoreInterests(profileID, interests)
// 		assert.EqualError(t, err, expectedErr.Error())
// 		mockDB.AssertExpectations(t)
// 		mockTx.AssertExpectations(t)
// 	})

// 	t.Run("failed to insert profile interest", func(t *testing.T) {
// 		mockDB := new(MockDB)
// 		repo := &repository.ProfileRepo{DB: mockDB}

// 		profileID := 1
// 		interests := []string{"music"}

// 		mockTx := new(MockDB)
// 		mockDB.On("Begin", context.Background()).Return(mockTx, nil)

// 		rows := &MockRows{
// 			data: [][]interface{}{{42}},
// 		}
// 		mockTx.On("QueryRow", context.Background(), GetInterestIdByDescription, []interface{}{"music"}).Return(rows)

// 		expectedErr := errors.New("profile interest insert failed")
// 		mockTx.On("Exec", context.Background(), InsertProfileInterest, []interface{}{profileID, 42}).Return(pgconn.NewCommandTag(""), expectedErr)

// 		mockTx.On("Rollback", context.Background()).Return(nil)

// 		err := repo.StoreInterests(profileID, interests)
// 		assert.EqualError(t, err, expectedErr.Error())
// 		mockDB.AssertExpectations(t)
// 		mockTx.AssertExpectations(t)
// 	})

// 	t.Run("failed to commit transaction", func(t *testing.T) {
// 		mockDB := new(MockDB)
// 		repo := &repository.ProfileRepo{DB: mockDB}

// 		profileID := 1
// 		interests := []string{"music"}

// 		mockTx := new(MockDB)
// 		mockDB.On("Begin", context.Background()).Return(mockTx, nil)

// 		rows := &MockRows{
// 			data: [][]interface{}{{42}},
// 		}
// 		mockTx.On("QueryRow", context.Background(), GetInterestIdByDescription, []interface{}{"music"}).Return(rows)

// 		mockTx.On("Exec", context.Background(), InsertProfileInterest, []interface{}{profileID, 42}).Return(pgconn.NewCommandTag("INSERT 0 1"), nil)

// 		expectedErr := errors.New("commit failed")
// 		mockTx.On("Commit", context.Background()).Return(expectedErr)
// 		mockTx.On("Rollback", context.Background()).Return(nil)

// 		err := repo.StoreInterests(profileID, interests)
// 		assert.EqualError(t, err, expectedErr.Error())
// 		mockDB.AssertExpectations(t)
// 		mockTx.AssertExpectations(t)
// 	})
// }

// func TestSQL_DeleteProfile(t *testing.T) {
// 	mockDB := new(MockDB)
// 	repo := &repository.ProfileRepo{DB: mockDB}

// 	mockDB.On("Query", mock.Anything, repository.FindUserProfileQuery, []interface{}{2}).
// 		Return(&MockRows{data: [][]interface{}{{5}}}, nil)

// 	mockDB.On("Exec", mock.Anything, repository.DeleteProfileQuery, []interface{}{5}).
// 		Return(pgconn.NewCommandTag("DELETE 1"), nil)

// 	err := repo.DeleteProfile(2)
// 	require.NoError(t, err)
// 	mockDB.AssertExpectations(t)
// }

// func ProfilesTestInitPostgresConfig(t *testing.T) {
// 	connStr := repository.InitPostgresConfig()
// 	expected := "host=localhost port=5432 user=test password=secret dbname=db sslmode=disable"
// 	assert.NotNil(t, expected, connStr)
// }

// func TestStorePhotos(t *testing.T) {
// 	mockDB := new(MockDB)
// 	repo := &repository.ProfileRepo{DB: mockDB}

// 	photos := []string{
// 		"photo1.jpg",
// 		"photo2.jpg",
// 	}

// 	for _, p := range photos {
// 		mockDB.On("Exec", mock.Anything,
// 			mock.MatchedBy(func(sql string) bool {
// 				return strings.Contains(sql, "INSERT INTO static")
// 			}),
// 			[]interface{}{1, p}).
// 			Return(pgconn.NewCommandTag("INSERT 0 1"), nil)
// 	}

// 	err := repo.StorePhotos(1, photos)
// 	assert.NoError(t, err)
// 	mockDB.AssertExpectations(t)
// }

// func TestSetLike(t *testing.T) {
// 	mockDB := new(MockDB)
// 	repo := &repository.ProfileRepo{DB: mockDB}

// 	fromID, toID, status := 1, 2, 1

// 	mockDB.On("Query", mock.Anything,
// 		mock.MatchedBy(func(sql string) bool {
// 			return strings.Contains(sql, "SELECT like_id, status FROM likes")
// 		}),
// 		[]interface{}{fromID, toID}).
// 		Return(&MockRows{data: [][]interface{}{}}, nil)

// 	mockDB.On("Query", mock.Anything,
// 		mock.MatchedBy(func(sql string) bool {
// 			return strings.Contains(sql, "INSERT INTO likes") &&
// 				strings.Contains(sql, "RETURNING like_id")
// 		}),
// 		[]interface{}{fromID, toID, status}).
// 		Return(&MockRows{data: [][]interface{}{{1}}}, nil)

// 	_, err := repo.SetLike(fromID, toID, status)
// 	assert.NoError(t, err)
// 	mockDB.AssertExpectations(t)
// }

// func TestGetPhotos(t *testing.T) {
// 	mockDB := new(MockDB)
// 	repo := &repository.ProfileRepo{DB: mockDB}

// 	userID := 1

// 	expectedSQL := `SELECT path FROM static WHERE profile_id = ( SELECT profile_id FROM users WHERE user_id = $1 )`
// 	mockDB.On("Query", mock.Anything, expectedSQL, []interface{}{userID}).
// 		Return(&MockRows{
// 			data: [][]interface{}{
// 				{"photo1.jpg"},
// 				{"photo2.jpg"},
// 			},
// 		}, nil)

// 	photos, err := repo.GetPhotos(userID)

// 	assert.NoError(t, err)
// 	assert.Len(t, photos, 2)
// 	assert.Equal(t, "photo1.jpg", photos[0])
// 	assert.Equal(t, "photo2.jpg", photos[1])
// 	mockDB.AssertExpectations(t)
// }

// func TestDeletePhoto(t *testing.T) {
// 	mockDB := new(MockDB)
// 	repo := &repository.ProfileRepo{DB: mockDB}

// 	profileID := 1
// 	photoPath := "image.png"

// 	expectedSQL := `DELETE FROM "static" WHERE profile_id = $1 AND path = $2`
// 	mockDB.On("Exec", mock.Anything, expectedSQL, []interface{}{profileID, "/" + photoPath}).
// 		Return(pgconn.NewCommandTag("DELETE 1"), nil)

// 	err := repo.DeletePhoto(profileID, photoPath)

// 	assert.NoError(t, err)
// 	mockDB.AssertExpectations(t)
// }

// func TestStorePhoto(t *testing.T) {
// 	mockDB := new(MockDB)
// 	repo := &repository.ProfileRepo{DB: mockDB}

// 	p := "image.jpg"
// 	userID := 0

// 	expectedSQL := `INSERT INTO static (profile_id, path, created_at, updated_at) VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING profile_id, path, created_at`
// 	mockDB.On("Exec", mock.Anything, expectedSQL, []interface{}{userID, p}).
// 		Return(pgconn.NewCommandTag("INSERT 1 1"), nil)

// 	err := repo.StorePhoto(userID, p)

// 	assert.NoError(t, err)
// 	mockDB.AssertExpectations(t)
// }

// func TestStoreProfile(t *testing.T) {
// 	mockDB := new(MockDB)
// 	repo := &repository.ProfileRepo{DB: mockDB}

// 	p := model.Profile{
// 		ProfileId:   1,
// 		FirstName:   "Иван",
// 		LastName:    "Иванов",
// 		IsMale:      true,
// 		Height:      180,
// 		Birthday:    time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
// 		Description: "Привет, я Иван",
// 		Location:    "Россия@Москва@Центральный",
// 	}

// 	locationParts := strings.Split(p.Location, "@")
// 	country, city, district := locationParts[0], locationParts[1], locationParts[2]

// 	mockDB.On("Query", mock.Anything,
// 		`INSERT INTO locations (country, city, district) VALUES ($1, $2, $3) RETURNING location_id`,
// 		[]interface{}{country, city, district}).
// 		Return(&MockRows{
// 			data: [][]interface{}{{1}},
// 		}, nil)

// 	mockDB.On("Query", mock.Anything,
// 		`INSERT INTO profiles (firstname, lastname, is_male, birthday, height, description, location_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING profile_id`,
// 		[]interface{}{
// 			p.FirstName,
// 			p.LastName,
// 			p.IsMale,
// 			p.Birthday,
// 			p.Height,
// 			p.Description,
// 			1, // location_id
// 		}).
// 		Return(&MockRows{
// 			data: [][]interface{}{{1}},
// 		}, nil)

// 	_, err := repo.StoreProfile(p)

// 	assert.NoError(t, err)
// 	mockDB.AssertExpectations(t)
// }

// func TestGetProfilesByUserId(t *testing.T) {
// 	mockDB := new(MockDB)
// 	repo := &repository.ProfileRepo{DB: mockDB}

// 	testTime := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
// 	expectedSQL := `SELECT p.profile_id, p.firstname, p.lastname, p.is_male, p.height, p.birthday, p.description, l.country, l.city, l.district, liked.profile_id AS liked_by_profile_id, s.path AS avatar, i.description AS interest, pr.preference_description, pr.preference_value FROM profiles p LEFT JOIN locations l ON p.location_id = l.location_id LEFT JOIN "static" s ON p.profile_id = s.profile_id LEFT JOIN profile_interests pi ON pi.profile_id = p.profile_id LEFT JOIN interests i ON pi.interest_id = i.interest_id LEFT JOIN profile_preferences pp ON pp.profile_id = p.profile_id LEFT JOIN preferences pr ON pp.preference_id = pr.preference_id LEFT JOIN likes liked ON liked.liked_profile_id = p.profile_id WHERE p.profile_id = $1`

// 	mockDB.On("Query", mock.Anything, expectedSQL, []interface{}{1}).
// 		Return(&MockRows{
// 			data: [][]interface{}{
// 				{
// 					1, "Иван", "Иванов", true, 180, testTime, "Описание",
// 					"Россия", "Москва", "Центральный",
// 					sql.NullInt64{Int64: 1, Valid: true},
// 					sql.NullString{String: "avatar.jpg", Valid: true},
// 					sql.NullString{String: "Технологии", Valid: true},
// 					sql.NullString{String: "Предпочтение 1", Valid: true},
// 					sql.NullString{String: "Высокое", Valid: true},
// 				},
// 			},
// 		}, nil)

// 	mockDB.On("Query", mock.Anything, expectedSQL, []interface{}{2}).
// 		Return(&MockRows{
// 			data: [][]interface{}{}, // Пустой результат
// 		}, nil)

// 	profiles, err := repo.GetProfilesByUserId(1)
// 	assert.NoError(t, err)
// 	assert.Len(t, profiles, 1)
// 	if len(profiles) > 0 {
// 		assert.Equal(t, 1, profiles[0].ProfileId)
// 		assert.Equal(t, "Иван", profiles[0].FirstName)
// 		assert.Equal(t, "Россия@Москва@Центральный", profiles[0].Location)
// 	}

// 	profiles, err = repo.GetProfilesByUserId(2)
// 	assert.NoError(t, err)
// 	assert.Len(t, profiles, 0)

// 	mockDB.AssertExpectations(t)
// }

// func TestUpdateProfile(t *testing.T) {
// 	mockDB := new(MockDB)
// 	repo := &repository.ProfileRepo{DB: mockDB}

// 	locationParts := strings.Split("Russia@Moscow@Centra", "@")
// 	mockDB.On("Query", mock.Anything, repository.GetLocationID,
// 		[]interface{}{locationParts[0], locationParts[1], locationParts[2]}).
// 		Return(&MockRows{
// 			data: [][]interface{}{},
// 		}, nil)

// 	mockDB.On("Exec", mock.Anything, repository.InsertLocation,
// 		[]interface{}{locationParts[0], locationParts[1], locationParts[2]}).
// 		Return(pgconn.NewCommandTag("INSERT 1"), nil)

// 	birthday := time.Date(1990, 5, 20, 0, 0, 0, 0, time.UTC)
// 	mockDB.On("Exec", mock.Anything, repository.UpdateProfileQuery,
// 		[]interface{}{"John", "Doe", true, 180, "Updated Description", 1, birthday, 100}).
// 		Return(pgconn.NewCommandTag("UPDATE 1"), nil)

// 	mockDB.On("Exec", mock.Anything, repository.DeleteProfileInterests,
// 		[]interface{}{100}).
// 		Return(pgconn.NewCommandTag("DELETE 1"), nil)

// 	mockDB.On("Query", mock.Anything, repository.GetInterestIdByDescription,
// 		[]interface{}{"Sport"}).
// 		Return(&MockRows{
// 			data: [][]interface{}{{1}},
// 		}, nil)

// 	mockDB.On("Exec", mock.Anything, repository.InsertProfileInterest,
// 		[]interface{}{100, 1}).
// 		Return(pgconn.NewCommandTag("INSERT 1"), nil)

// 	mockDB.On("Commit", mock.Anything).Return(nil)

// 	newProfile := model.Profile{
// 		FirstName:   "John",
// 		LastName:    "Doe",
// 		IsMale:      true,
// 		Height:      180,
// 		Description: "Updated Description",
// 		Location:    "Russia@Moscow@Centra",
// 		Interests:   []string{"Sport"},
// 		Birthday:    birthday,
// 	}

// 	err := repo.UpdateProfile(100, newProfile)
// 	require.NoError(t, err)

// 	mockDB.AssertExpectations(t)
// }
