package tests

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/png"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/repository"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	pgxmock "github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// func TestGetProfileById(t *testing.T) {
// 	tests := []struct {
// 		name        string
// 		profileID   int
// 		mockRows    func(mock sqlmock.Sqlmock)
// 		expected    model.Profile
// 		expectedErr bool
// 	}{
// 		{
// 			name:      "valid profile with all fields",
// 			profileID: 1,
// 			mockRows: func(mock sqlmock.Sqlmock) {
// 				mock.ExpectQuery(regexp.QuoteMeta(repository.GetProfileByIdQuery)).
// 					WithArgs(1).
// 					WillReturnRows(sqlmock.NewRows([]string{
// 						"profile_id", "firstname", "lastname", "is_male", "height",
// 						"birthday", "description", "country", "city", "district",
// 						"liked_by_profile_id", "avatar", "interest", "preference_description", "preference_value",
// 					}).AddRow(
// 						1, "Иван", "Иванов", true, 180,
// 						time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC), "Описание", "Россия", "Москва", "ЦАО",
// 						sql.NullInt64{Int64: 2, Valid: true},
// 						sql.NullString{String: "avatar.jpg", Valid: true},
// 						sql.NullString{String: "Музыка", Valid: true},
// 						sql.NullString{String: "Рост", Valid: true},
// 						sql.NullString{String: "180+", Valid: true},
// 					))
// 			},
// 			expected: model.Profile{
// 				ProfileId:   1,
// 				FirstName:   "Иван",
// 				LastName:    "Иванов",
// 				IsMale:      true,
// 				Height:      180,
// 				Birthday:    time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
// 				Description: "Описание",
// 				Location:    "Россия@Москва@ЦАО",
// 				LikedBy:     []int{2},
// 				Photos:      []string{"avatar.jpg"},
// 				Interests:   []string{"Музыка"},
// 				Preferences: []model.Preference{
// 					{Description: "Рост", Value: "180+"},
// 				},
// 			},
// 			expectedErr: false,
// 		},
// 		{
// 			name:      "empty profile (no data)",
// 			profileID: 42,
// 			mockRows: func(mock sqlmock.Sqlmock) {
// 				mock.ExpectQuery(regexp.QuoteMeta(repository.GetProfileByIdQuery)).
// 					WithArgs(42).
// 					WillReturnRows(sqlmock.NewRows([]string{
// 						"profile_id", "firstname", "lastname", "is_male", "height",
// 						"birthday", "description", "country", "city", "district",
// 						"liked_by_profile_id", "avatar", "interest", "preference_description", "preference_value",
// 					}))
// 			},
// 			expected:    model.Profile{},
// 			expectedErr: false,
// 		},
// 		{
// 			name:      "query error",
// 			profileID: 99,
// 			mockRows: func(mock sqlmock.Sqlmock) {
// 				mock.ExpectQuery(regexp.QuoteMeta(repository.GetProfileByIdQuery)).
// 					WithArgs(99).
// 					WillReturnError(sql.ErrConnDone)
// 			},
// 			expected:    model.Profile{},
// 			expectedErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			db, mock, err := sqlmock.New()
// 			assert.NoError(t, err)
// 			defer db.Close()

// 			tt.mockRows(mock)

// 			repo := &repository.ProfileRepo{DB: db}
// 			profile, err := repo.GetProfileById(tt.profileID)

// 			if tt.expectedErr {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.Equal(t, tt.expected.ProfileId, profile.ProfileId)
// 				assert.Equal(t, tt.expected.FirstName, profile.FirstName)
// 				assert.Equal(t, tt.expected.LastName, profile.LastName)
// 				assert.Equal(t, tt.expected.IsMale, profile.IsMale)
// 				assert.Equal(t, tt.expected.Height, profile.Height)
// 				assert.Equal(t, tt.expected.Birthday, profile.Birthday)
// 				assert.Equal(t, tt.expected.Description, profile.Description)
// 				assert.Equal(t, tt.expected.Location, profile.Location)
// 				assert.ElementsMatch(t, tt.expected.LikedBy, profile.LikedBy)
// 				assert.ElementsMatch(t, tt.expected.Photos, profile.Photos)
// 				assert.ElementsMatch(t, tt.expected.Interests, profile.Interests)
// 				assert.ElementsMatch(t, tt.expected.Preferences, profile.Preferences)
// 			}
// 			assert.NoError(t, mock.ExpectationsWereMet())
// 		})
// 	}
// }

func TestGetProfileById(t *testing.T) {
	tests := []struct {
		name        string
		profileID   int
		mockRows    func(mock pgxmock.PgxPoolIface)
		expected    model.Profile
		expectedErr bool
	}{
		{
			name:      "valid profile with all fields",
			profileID: 1,
			mockRows: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"profile_id", "firstname", "lastname", "is_male", "height",
					"birthday", "description", "country", "city", "district",
					"liked_by_profile_id", "avatar", "interest", "preference_description", "preference_value",
				}).AddRow(
					1, "Иван", "Иванов", true, 180,
					time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC), "Описание", "Россия", "Москва", "ЦАО",
					// pgtype.Int8{Int: 2, Valid: true},
					pgtype.Int8{Int: 2, Status: pgtype.Present},
					pgtype.Text{String: "avatar.jpg", Status: pgtype.Present},
					pgtype.Text{String: "Музыка", Status: pgtype.Present},
					pgtype.Text{String: "Рост", Status: pgtype.Present},
					pgtype.Text{String: "180+", Status: pgtype.Present},
				)
				mock.ExpectQuery(regexp.QuoteMeta(repository.GetProfileByIdQuery)).
					WithArgs(1).
					WillReturnRows(rows)
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
			mockRows: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"profile_id", "firstname", "lastname", "is_male", "height",
					"birthday", "description", "country", "city", "district",
					"liked_by_profile_id", "avatar", "interest", "preference_description", "preference_value",
				})
				mock.ExpectQuery(regexp.QuoteMeta(repository.GetProfileByIdQuery)).
					WithArgs(42).
					WillReturnRows(rows)
			},
			expected:    model.Profile{},
			expectedErr: false,
		},
		{
			name:      "query error",
			profileID: 99,
			mockRows: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(regexp.QuoteMeta(repository.GetProfileByIdQuery)).
					WithArgs(99).
					WillReturnError(errors.New("query error"))
			},
			expected:    model.Profile{},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			assert.NoError(t, err)
			defer mock.Close()

			tt.mockRows(mock)

			repo := &repository.ProfileRepo{DB: mock}
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

// func TestSQL_StoreInterests(t *testing.T) {
// 	db, mock, err := sqlmock.New()
// 	require.NoError(t, err)
// 	defer db.Close()

// 	repo := &repository.ProfileRepo{DB: db}

// 	mock.ExpectBegin()

// 	mock.ExpectQuery(`SELECT interest_id FROM interests WHERE description = \$1`).
// 		WithArgs("Музыка").
// 		WillReturnError(sql.ErrNoRows)

// 	mock.ExpectQuery(`INSERT INTO interests \(description\) VALUES \(\$1\) RETURNING interest_id`).
// 		WithArgs("Музыка").
// 		WillReturnRows(sqlmock.NewRows([]string{"interest_id"}).AddRow(3))

// 	mock.ExpectExec(`INSERT INTO profile_interests \(profile_id, interest_id\) VALUES \(\$1, \$2\)`).
// 		WithArgs(1, 3).
// 		WillReturnResult(sqlmock.NewResult(1, 1))

// 	mock.ExpectCommit()

// 	err = repo.StoreInterests(1, []string{"Музыка"})
// 	require.NoError(t, err)
// 	require.NoError(t, mock.ExpectationsWereMet())
// }

func TestSQL_StoreInterests(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := &repository.ProfileRepo{DB: mockPool}

	mockPool.ExpectBegin()

	mockPool.ExpectQuery(`SELECT interest_id FROM interests WHERE description = \$1`).
		WithArgs("Музыка").
		WillReturnError(pgx.ErrNoRows)

	mockPool.ExpectQuery(`INSERT INTO interests \(description\) VALUES \(\$1\) RETURNING interest_id`).
		WithArgs("Музыка").
		WillReturnRows(pgxmock.NewRows([]string{"interest_id"}).AddRow(3))

	mockPool.ExpectExec(`INSERT INTO profile_interests \(profile_id, interest_id\) VALUES \(\$1, \$2\)`).
		WithArgs(1, 3).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mockPool.ExpectCommit()

	err = repo.StoreInterests(1, []string{"Музыка"})
	require.NoError(t, err)
	require.NoError(t, mockPool.ExpectationsWereMet())
}

// func TestSQL_DeleteProfile(t *testing.T) {
// 	db, mock, err := sqlmock.New()
// 	require.NoError(t, err)
// 	defer db.Close()

// 	repo := &repository.ProfileRepo{DB: db}

// 	mock.ExpectQuery(repository.FindUserProfileQuery).WithArgs(2).
// 		WillReturnRows(sqlmock.NewRows([]string{"profile_id"}).AddRow(5))

// 	mock.ExpectExec(repository.DeleteProfileQuery).WithArgs(5).
// 		WillReturnResult(sqlmock.NewResult(1, 1))

// 	repo.DeleteProfile(2)
// }

func TestSQL_DeleteProfile(t *testing.T) {
	// Создаем mock pool
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	// Инициализируем репозиторий с mock pool
	repo := &repository.ProfileRepo{DB: mockPool}

	// Ожидаем запрос на поиск профиля пользователя
	mockPool.ExpectQuery(regexp.QuoteMeta(repository.FindUserProfileQuery)).
		WithArgs(2).
		WillReturnRows(pgxmock.NewRows([]string{"profile_id"}).AddRow(5))

	// Ожидаем запрос на удаление профиля
	mockPool.ExpectExec(regexp.QuoteMeta(repository.DeleteProfileQuery)).
		WithArgs(5).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	// Вызываем тестируемый метод
	err = repo.DeleteProfile(2)
	require.NoError(t, err)
	require.NoError(t, mockPool.ExpectationsWereMet())
}

func TestNewUserRepo(t *testing.T) {
	repo, _ := repository.NewUserRepo()
	assert.NotNil(t, repo)
}

func ProfilesTestInitPostgresConfig(t *testing.T) {
	connStr := repository.InitPostgresConfig()
	expected := "host=localhost port=5432 user=test password=secret dbname=db sslmode=disable"
	assert.NotNil(t, expected, connStr)
}

// func TestStorePhotos(t *testing.T) {
// 	db, mock, _ := sqlmock.New()
// 	defer db.Close()
// 	repo := &repository.ProfileRepo{DB: db}

// 	photos := []string{
// 		"photo1.jpg",
// 		"photo2.jpg",
// 	}

// 	for _, p := range photos {
// 		mock.ExpectExec(`INSERT INTO static`).
// 			WithArgs(1, p).
// 			WillReturnResult(sqlmock.NewResult(1, 1))
// 	}

// 	err := repo.StorePhotos(1, photos)
// 	assert.NoError(t, err)
// 	assert.NoError(t, mock.ExpectationsWereMet())
// }

func TestStorePhotos(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := &repository.ProfileRepo{DB: mockPool}

	photos := []string{
		"photo1.jpg",
		"photo2.jpg",
	}

	for _, p := range photos {
		mockPool.ExpectExec(regexp.QuoteMeta(`INSERT INTO static`)).
			WithArgs(1, p).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
	}

	err = repo.StorePhotos(1, photos)
	require.NoError(t, err)
	require.NoError(t, mockPool.ExpectationsWereMet())
}

// func TestSetLike(t *testing.T) {
// 	db, mock, _ := sqlmock.New()
// 	defer db.Close()
// 	repo := &repository.ProfileRepo{DB: db}

// 	fromID, toID, status := 1, 2, 1

// 	mock.ExpectQuery(`SELECT like_id, status FROM likes`).
// 		WithArgs(fromID, toID).
// 		WillReturnRows(sqlmock.NewRows([]string{}))

// 	mock.ExpectQuery(`INSERT INTO likes .* RETURNING like_id`).
// 		WithArgs(fromID, toID, status).
// 		WillReturnRows(sqlmock.NewRows([]string{"like_id"}).AddRow(1))

// 	_, err := repo.SetLike(fromID, toID, status)
// 	assert.NoError(t, err)
// 	assert.NoError(t, mock.ExpectationsWereMet())
// }

func TestSetLike(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := &repository.ProfileRepo{DB: mockPool}

	fromID, toID, status := 1, 2, 1

	mockPool.ExpectQuery(regexp.QuoteMeta(`SELECT like_id, status FROM likes`)).
		WithArgs(fromID, toID).
		WillReturnRows(pgxmock.NewRows([]string{}))

	mockPool.ExpectQuery(regexp.QuoteMeta(`INSERT INTO likes .* RETURNING like_id`)).
		WithArgs(fromID, toID, status).
		WillReturnRows(pgxmock.NewRows([]string{"like_id"}).AddRow(1))

	_, err = repo.SetLike(fromID, toID, status)
	require.NoError(t, err)
	require.NoError(t, mockPool.ExpectationsWereMet())
}

// func TestGetPhotos(t *testing.T) {
// 	db, mock, _ := sqlmock.New()
// 	defer db.Close()
// 	repo := &repository.ProfileRepo{DB: db}

// 	userID := 1

// 	mock.ExpectQuery(`SELECT path FROM static WHERE profile_id = \( SELECT profile_id FROM users WHERE user_id = \$1 \)`).
// 		WithArgs(userID).
// 		WillReturnRows(sqlmock.NewRows([]string{"path"}).
// 			AddRow("photo1.jpg").
// 			AddRow("photo2.jpg"))

// 	photos, err := repo.GetPhotos(userID)
// 	assert.NoError(t, err)
// 	assert.Len(t, photos, 2)
// 	assert.Equal(t, "photo1.jpg", photos[0])
// 	assert.Equal(t, "photo2.jpg", photos[1])
// 	assert.NoError(t, mock.ExpectationsWereMet())
// }

func TestGetPhotos(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := &repository.ProfileRepo{DB: mockPool}

	userID := 1

	mockPool.ExpectQuery(regexp.QuoteMeta(
		`SELECT path FROM static WHERE profile_id = ( SELECT profile_id FROM users WHERE user_id = $1 )`,
	)).
		WithArgs(userID).
		WillReturnRows(
			pgxmock.NewRows([]string{"path"}).
				AddRow("photo1.jpg").
				AddRow("photo2.jpg"),
		)

	photos, err := repo.GetPhotos(userID)

	require.NoError(t, err)
	require.Len(t, photos, 2)
	require.Equal(t, "photo1.jpg", photos[0])
	require.Equal(t, "photo2.jpg", photos[1])
	require.NoError(t, mockPool.ExpectationsWereMet())
}

// func TestDeletePhoto(t *testing.T) {
// 	db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("Не удалось создать mock для базы данных: %v", err)
// 	}
// 	defer db.Close()

// 	repo := &repository.ProfileRepo{DB: db}

// 	mock.ExpectExec(`DELETE FROM "static" WHERE profile_id = \$1 AND path = \$2`).
// 		WithArgs(1, "/image.png").
// 		WillReturnResult(sqlmock.NewResult(0, 1))

// 	err = repo.DeletePhoto(1, "image.png")

// 	assert.NoError(t, err)
// 	assert.NoError(t, mock.ExpectationsWereMet())
// }

func TestDeletePhoto(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Не удалось создать mock для базы данных: %v", err)
	}
	defer mockPool.Close()

	repo := &repository.ProfileRepo{DB: mockPool}

	mockPool.ExpectExec(regexp.QuoteMeta(
		`DELETE FROM "static" WHERE profile_id = $1 AND path = $2`,
	)).
		WithArgs(1, "/image.png").
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.DeletePhoto(1, "image.png")

	require.NoError(t, err)
	require.NoError(t, mockPool.ExpectationsWereMet())
}

// func TestStorePhoto(t *testing.T) {
// 	db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("Не удалось создать mock для базы данных: %v", err)
// 	}
// 	defer db.Close()

// 	repo := &repository.ProfileRepo{DB: db}

// 	p := "image.jpg"
// 	userID := 0

// 	mock.ExpectExec(`^INSERT INTO static \(.+\) VALUES \(\$1, \$2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP\) RETURNING profile_id, path, created_at;`).
// 		WithArgs(userID, p).
// 		WillReturnResult(sqlmock.NewResult(1, 1))

// 	err = repo.StorePhoto(userID, p)

// 	assert.NoError(t, err)
// 	assert.NoError(t, mock.ExpectationsWereMet())
// }

func TestStorePhoto(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Не удалось создать mock для базы данных: %v", err)
	}
	defer mockPool.Close()

	repo := &repository.ProfileRepo{DB: mockPool}

	p := "image.jpg"
	userID := 0

	mockPool.ExpectExec(regexp.QuoteMeta(
		`INSERT INTO static (profile_id, path, created_at, updated_at) VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
	)).
		WithArgs(userID, p).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.StorePhoto(userID, p)

	require.NoError(t, err)
	require.NoError(t, mockPool.ExpectationsWereMet())
}

// func TestStoreProfile(t *testing.T) {
// 	db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("Не удалось создать mock для базы данных: %v", err)
// 	}
// 	defer db.Close()

// 	repo := &repository.ProfileRepo{DB: db}

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

// 	mock.ExpectQuery(`INSERT INTO locations \(country, city, district\) VALUES \(\$1, \$2, \$3\) RETURNING location_id`).
// 		WithArgs("Россия", "Москва", "Центральный").
// 		WillReturnRows(sqlmock.NewRows([]string{"location_id"}).AddRow(1))

// 	mock.ExpectQuery(`INSERT INTO profiles \(firstname, lastname, is_male, birthday, height, description, location_id, created_at, updated_at\)`).
// 		WithArgs(
// 			p.FirstName,
// 			p.LastName,
// 			p.IsMale,
// 			p.Birthday,
// 			p.Height,
// 			p.Description,
// 			1,
// 		).
// 		WillReturnRows(sqlmock.NewRows([]string{"profile_id"}).AddRow(1))

// 	_, err = repo.StoreProfile(p)

// 	assert.NoError(t, err)
// 	assert.NoError(t, mock.ExpectationsWereMet())
// }

func TestStoreProfile(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Не удалось создать mock для базы данных: %v", err)
	}
	defer mockPool.Close()

	repo := &repository.ProfileRepo{DB: mockPool}

	p := model.Profile{
		ProfileId:   1,
		FirstName:   "Иван",
		LastName:    "Иванов",
		IsMale:      true,
		Height:      180,
		Birthday:    time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Description: "Привет, я Иван",
		Location:    "Россия@Москва@Центральный",
	}

	locationParts := strings.Split(p.Location, "@")
	if len(locationParts) != 3 {
		t.Fatal("Неверный формат локации")
	}

	mockPool.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO locations (country, city, district) VALUES ($1, $2, $3) RETURNING location_id`,
	)).
		WithArgs(locationParts[0], locationParts[1], locationParts[2]).
		WillReturnRows(pgxmock.NewRows([]string{"location_id"}).AddRow(1))

	mockPool.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO profiles (firstname, lastname, is_male, birthday, height, description, location_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING profile_id`,
	)).
		WithArgs(
			p.FirstName,
			p.LastName,
			p.IsMale,
			p.Birthday,
			p.Height,
			p.Description,
			1, // location_id
		).
		WillReturnRows(pgxmock.NewRows([]string{"profile_id"}).AddRow(1))

	_, err = repo.StoreProfile(p)

	require.NoError(t, err)
	require.NoError(t, mockPool.ExpectationsWereMet())
}

// func TestGetProfilesByUserId(t *testing.T) {
// 	db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("Не удалось создать mock для базы данных: %v", err)
// 	}
// 	defer db.Close()

// 	repo := &repository.ProfileRepo{DB: db}

// 	rows := sqlmock.NewRows([]string{"profile_id", "firstname", "lastname", "is_male", "height", "birthday", "description", "country", "city", "district", "liked_by_profile_id", "avatar", "interest", "preference_description", "preference_value"}).
// 		AddRow(1, "Иван", "Иванов", true, 180, time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC), "Описание", "Россия", "Москва", "Центральный", 1, "avatar.jpg", "Технологии", "Предпочтение 1", "Высокое")

// 	mock.ExpectQuery(`SELECT p.profile_id, p.firstname, p.lastname, p.is_male, p.height, p.birthday, p.description, l.country, l.city, l.district, liked.profile_id AS liked_by_profile_id, s.path AS avatar, i.description AS interest, pr.preference_description, pr.preference_value FROM profiles p LEFT JOIN locations l ON p.location_id = l.location_id LEFT JOIN "static" s ON p.profile_id = s.profile_id LEFT JOIN profile_interests pi ON pi.profile_id = p.profile_id LEFT JOIN interests i ON pi.interest_id = i.interest_id LEFT JOIN profile_preferences pp ON pp.profile_id = p.profile_id LEFT JOIN preferences pr ON pp.preference_id = pr.preference_id LEFT JOIN likes liked ON liked.liked_profile_id = p.profile_id WHERE p.profile_id = \$1`).
// 		WithArgs(1).
// 		WillReturnRows(rows)

// 	repo.GetProfilesByUserId(0)

// 	mock.ExpectQuery(`SELECT .* FROM profiles .* WHERE p.profile_id = \$1`).
// 		WithArgs(2).
// 		WillReturnRows(rows)

// 	profiles, err := repo.GetProfilesByUserId(1)
// 	assert.NoError(t, err)
// 	assert.Len(t, profiles, 0)
// 	assert.NoError(t, mock.ExpectationsWereMet())
// }

func TestGetProfilesByUserId(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Не удалось создать mock для базы данных: %v", err)
	}
	defer mockPool.Close()

	repo := &repository.ProfileRepo{DB: mockPool}

	testBirthday := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	rows := pgxmock.NewRows([]string{
		"profile_id", "firstname", "lastname", "is_male", "height",
		"birthday", "description", "country", "city", "district",
		"liked_by_profile_id", "avatar", "interest", "preference_description",
		"preference_value",
	}).AddRow(
		1, "Иван", "Иванов", true, 180,
		testBirthday, "Описание", "Россия", "Москва", "Центральный",
		pgtype.Int8{Int: 1, Status: pgtype.Present},
		pgtype.Text{String: "avatar.jpg", Status: pgtype.Present},
		pgtype.Text{String: "Технологии", Status: pgtype.Present},
		pgtype.Text{String: "Предпочтение 1", Status: pgtype.Present},
		pgtype.Text{String: "Высокое", Status: pgtype.Present},
	)

	query := `SELECT p.profile_id, p.firstname, p.lastname, p.is_male, p.height, 
              p.birthday, p.description, l.country, l.city, l.district, 
              liked.profile_id AS liked_by_profile_id, s.path AS avatar, 
              i.description AS interest, pr.preference_description, 
              pr.preference_value 
              FROM profiles p 
              LEFT JOIN locations l ON p.location_id = l.location_id 
              LEFT JOIN "static" s ON p.profile_id = s.profile_id 
              LEFT JOIN profile_interests pi ON pi.profile_id = p.profile_id 
              LEFT JOIN interests i ON pi.interest_id = i.interest_id 
              LEFT JOIN profile_preferences pp ON pp.profile_id = p.profile_id 
              LEFT JOIN preferences pr ON pp.preference_id = pr.preference_id 
              LEFT JOIN likes liked ON liked.liked_profile_id = p.profile_id 
              WHERE p.profile_id = \$1`

	mockPool.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(1).
		WillReturnRows(rows)

	profiles, err := repo.GetProfilesByUserId(1)

	require.NoError(t, err)
	require.Len(t, profiles, 1) // Изменил ожидаемое количество профилей с 0 на 1

	if len(profiles) > 0 {
		profile := profiles[0]
		require.Equal(t, 1, profile.ProfileId)
		require.Equal(t, "Иван", profile.FirstName)
		require.Equal(t, "Иванов", profile.LastName)
		require.Equal(t, true, profile.IsMale)
		require.Equal(t, 180, profile.Height)
		require.Equal(t, testBirthday, profile.Birthday)
		require.Equal(t, "Описание", profile.Description)
		require.Equal(t, "Россия@Москва@Центральный", profile.Location)
		require.Contains(t, profile.LikedBy, 1)
		require.Contains(t, profile.Photos, "avatar.jpg")
		require.Contains(t, profile.Interests, "Технологии")
		require.Contains(t, profile.Preferences, model.Preference{
			Description: "Предпочтение 1",
			Value:       "Высокое",
		})
	}

	require.NoError(t, mockPool.ExpectationsWereMet())
}

func setupMinioClient(t *testing.T) *repository.StaticRepo {
	endpoint := "localhost:9000"
	accessKeyID := "minioadmin"
	secretAccessKey := "miniopassword"
	useSSL := false

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	require.NoError(t, err)

	bucketName := "test-bucket"

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucketName)
	require.NoError(t, err)

	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		require.NoError(t, err)
	}

	return repository.NewStaticRepoCl(client, bucketName)
}

func TestUploadAndGetImage(t *testing.T) {
	repo := setupMinioClient(t)

	filename := "test-image.png"
	contentType := "image/png"

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	require.NoError(t, err)

	err = repo.UploadImage(buf.Bytes(), filename, contentType)
	require.NoError(t, err)

	data, err := repo.GetImages([]string{filename})
	require.NoError(t, err)
	require.Len(t, data, 1)

	imgDecoded, _, err := image.Decode(bytes.NewReader(data[0]))
	require.NoError(t, err)
	assert.Equal(t, 100, imgDecoded.Bounds().Dx())
	assert.Equal(t, 100, imgDecoded.Bounds().Dy())
}

func TestDeleteImage(t *testing.T) {
	repo := setupMinioClient(t)

	filename := "to-delete.png"
	contentType := "image/png"

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	require.NoError(t, err)

	err = repo.UploadImage(buf.Bytes(), filename, contentType)
	require.NoError(t, err)

	err = repo.DeleteImage(0, filename)
	require.NoError(t, err)

	_, err = repo.GetImages([]string{filename})
	require.Error(t, err)
}

func TestGenerateImagePNG(t *testing.T) {
	repo := setupMinioClient(t)

	imgBytes, err := repo.GenerateImage("image/png", true)
	require.NoError(t, err)

	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	require.NoError(t, err)

	assert.True(t, img.Bounds().Dx() > 0)
	assert.True(t, img.Bounds().Dy() > 0)
}

func TestGenerateImageJPEG(t *testing.T) {
	repo := setupMinioClient(t)

	imgBytes, err := repo.GenerateImage("image/jpeg", false)
	require.NoError(t, err)

	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	require.NoError(t, err)

	assert.True(t, img.Bounds().Dx() > 0)
	assert.True(t, img.Bounds().Dy() > 0)
}

// func TestSQL_UpdateProfile(t *testing.T) {
// 	db, mock, err := sqlmock.New()
// 	require.NoError(t, err)
// 	defer db.Close()

// 	repo := &repository.ProfileRepo{DB: db}

// 	mock.ExpectBegin()

// 	mock.ExpectQuery(repository.GetLocationID).
// 		WithArgs("Russia", "Moscow", "Centra").
// 		WillReturnRows(sqlmock.NewRows([]string{"location_id"}).AddRow(0))

// 	mock.ExpectExec(repository.InsertLocation).
// 		WithArgs("Russia", "Moscow", "Centra").
// 		WillReturnResult(sqlmock.NewResult(1, 1))

// 	mock.ExpectExec(repository.UpdateProfileQuery).
// 		WithArgs("John", "Doe", true, 180, "Updated Description", 1, time.Date(1990, 5, 20, 0, 0, 0, 0, time.UTC), 100).
// 		WillReturnResult(sqlmock.NewResult(1, 1))

// 	mock.ExpectExec(repository.DeleteProfileInterests).
// 		WithArgs(100).
// 		WillReturnResult(sqlmock.NewResult(1, 1))

// 	mock.ExpectQuery(repository.GetInterestIdByDescription).
// 		WithArgs("Sport").
// 		WillReturnRows(sqlmock.NewRows([]string{"interest_id"}).AddRow(1))

// 	mock.ExpectExec(repository.InsertProfileInterest).
// 		WithArgs(100, 1).
// 		WillReturnResult(sqlmock.NewResult(1, 1))

// 	mock.ExpectCommit()

// 	newProfile := model.Profile{
// 		FirstName:   "John",
// 		LastName:    "Doe",
// 		IsMale:      true,
// 		Height:      180,
// 		Description: "Updated Description",
// 		Location:    "Russia@Moscow@Centra",
// 		Interests:   []string{"Sport"},
// 	}

// 	repo.UpdateProfile(100, newProfile)

// }

func TestSQL_UpdateProfile(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	repo := &repository.ProfileRepo{DB: mockPool}

	mockPool.ExpectBegin()

	mockPool.ExpectQuery(regexp.QuoteMeta(repository.GetLocationID)).
		WithArgs("Russia", "Moscow", "Centra").
		WillReturnRows(pgxmock.NewRows([]string{"location_id"}).AddRow(0))

	mockPool.ExpectExec(regexp.QuoteMeta(repository.InsertLocation)).
		WithArgs("Russia", "Moscow", "Centra").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mockPool.ExpectExec(regexp.QuoteMeta(repository.UpdateProfileQuery)).
		WithArgs(
			"John",
			"Doe",
			true,
			180,
			"Updated Description",
			1, // location_id
			time.Date(1990, 5, 20, 0, 0, 0, 0, time.UTC),
			100, // profile_id
		).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mockPool.ExpectExec(regexp.QuoteMeta(repository.DeleteProfileInterests)).
		WithArgs(100).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	mockPool.ExpectQuery(regexp.QuoteMeta(repository.GetInterestIdByDescription)).
		WithArgs("Sport").
		WillReturnRows(pgxmock.NewRows([]string{"interest_id"}).AddRow(1))

	mockPool.ExpectExec(regexp.QuoteMeta(repository.InsertProfileInterest)).
		WithArgs(100, 1).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mockPool.ExpectCommit()

	newProfile := model.Profile{
		FirstName:   "John",
		LastName:    "Doe",
		IsMale:      true,
		Height:      180,
		Description: "Updated Description",
		Location:    "Russia@Moscow@Centra",
		Interests:   []string{"Sport"},
		Birthday:    time.Date(1990, 5, 20, 0, 0, 0, 0, time.UTC),
	}

	err = repo.UpdateProfile(100, newProfile)
	require.NoError(t, err)
	require.NoError(t, mockPool.ExpectationsWereMet())
}
