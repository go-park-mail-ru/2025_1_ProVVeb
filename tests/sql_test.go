package tests

import (
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"
	"time"

	"context"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/stretchr/testify/require"
)

type sqlTestCase_User struct {
	name           string
	query          string
	args           []driver.Value
	expectedRows   []string
	returnedValues []driver.Value
	expectedResult model.User
	expectErr      bool
}

func TestSQL_GetUserByLogins(t *testing.T) {
	tests := []sqlTestCase_User{
		{
			name:  "GetUserByLogin success",
			query: repository.GetUserByLoginQuery,
			args:  []driver.Value{"testuser"},
			expectedRows: []string{
				"user_id", "login", "email", "password", "phone", "status",
			},
			returnedValues: []driver.Value{
				1, "testuser", "test@example.com", "hashed_password", "+1234567890", 1,
			},
			expectedResult: model.User{
				UserId:   1,
				Login:    "testuser",
				Email:    "test@example.com",
				Password: "hashed_password",
				Phone:    "+1234567890",
				Status:   1,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			conn, err := db.Conn(context.Background())
			require.NoError(t, err)
			defer conn.Close()

			repo := &repository.UserRepo{DB: conn}

			rows := sqlmock.NewRows(tt.expectedRows).AddRow(tt.returnedValues...)

			mock.ExpectQuery(regexp.QuoteMeta(tt.query)).
				WithArgs(tt.args...).
				WillReturnRows(rows)

			user, err := repo.GetUserByLogin(context.Background(), tt.args[0].(string))

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedResult, user)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSQL_GetMatches(t *testing.T) {
	tests := []struct {
		name             string
		userID           int
		matchRows        [][]driver.Value
		profileRows      [][]driver.Value
		expectedProfiles []model.Profile
		expectErr        bool
	}{
		{
			name:   "successfully returns matched profiles",
			userID: 1,
			matchRows: [][]driver.Value{
				{1, 2},
				{2, 1},
			},
			profileRows: [][]driver.Value{
				{2, "Alice", "Smith", true, 170, time.Date(1995, 5, 10, 0, 0, 0, 0, time.UTC), "desc", "Russia", nil, "/photo.jpg", "music", "bodyType:slim"},
				{2, "Alice", "Smith", true, 170, time.Date(1995, 5, 10, 0, 0, 0, 0, time.UTC), "desc", "Russia", nil, "/photo.jpg", "travel", "hairColor:blonde"},
			},
			expectedProfiles: []model.Profile{
				{
					ProfileId:   2,
					FirstName:   "Alice",
					LastName:    "Smith",
					IsMale:      true,
					Height:      170,
					Birthday:    time.Date(1995, 5, 10, 0, 0, 0, 0, time.UTC),
					Description: "desc",
					Location:    "Russia",
					Avatar:      "http://213.219.214.83:8080/static/avatars/photo.jpg",
					Card:        "http://213.219.214.83:8080/static/cards/photo.jpg",
					Interests:   []string{"music", "travel"},
					Preferences: []string{"bodyType:slim", "hairColor:blonde"},
				},
				{
					ProfileId:   2,
					FirstName:   "Alice",
					LastName:    "Smith",
					IsMale:      true,
					Height:      170,
					Birthday:    time.Date(1995, 5, 10, 0, 0, 0, 0, time.UTC),
					Description: "desc",
					Location:    "Russia",
					Avatar:      "http://213.219.214.83:8080/static/avatars/photo.jpg",
					Card:        "http://213.219.214.83:8080/static/cards/photo.jpg",
					Interests:   []string{"music", "travel"},
					Preferences: []string{"bodyType:slim", "hairColor:blonde"},
				},
			},
			expectErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			conn, err := db.Conn(context.Background())
			require.NoError(t, err)
			defer conn.Close()

			repo := &repository.UserRepo{DB: conn}

			matchRows := sqlmock.NewRows([]string{"profile_id", "matched_profile_id"})
			for _, r := range tt.matchRows {
				matchRows.AddRow(r...)
			}
			mock.ExpectQuery(regexp.QuoteMeta(repository.GetMatches)).
				WithArgs(tt.userID).
				WillReturnRows(matchRows)

			for range tt.matchRows {
				rows := sqlmock.NewRows([]string{
					"profile_id", "firstname", "lastname", "is_male", "height", "birthday", "description",
					"country", "liked_by_profile_id", "avatar", "interest", "preference",
				})
				for _, r := range tt.profileRows {
					rows.AddRow(r...)
				}
				mock.ExpectQuery(regexp.QuoteMeta(repository.GetProfileByIdQuery)).
					WithArgs(2).
					WillReturnRows(rows)
			}

			profiles, err := repo.GetMatches(tt.userID)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedProfiles, profiles)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSQL_StoreProfile(t *testing.T) {
	tests := []struct {
		name         string
		inputProfile model.Profile
		returnedID   int
		expectErr    bool
	}{
		{
			name: "successfully stores profile",
			inputProfile: model.Profile{
				FirstName:   "Alice",
				LastName:    "Smith",
				IsMale:      true,
				Birthday:    time.Date(1995, 5, 5, 0, 0, 0, 0, time.UTC),
				Height:      165,
				Description: "Описание профиля",
			},
			returnedID: 42,
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			conn, err := db.Conn(context.Background())
			require.NoError(t, err)
			defer conn.Close()

			repo := &repository.UserRepo{DB: conn}

			mock.ExpectQuery(regexp.QuoteMeta(repository.CreateProfileQuery)).
				WithArgs(tt.inputProfile.FirstName, tt.inputProfile.LastName, tt.inputProfile.IsMale,
					tt.inputProfile.Birthday, tt.inputProfile.Height, tt.inputProfile.Description).
				WillReturnRows(sqlmock.NewRows([]string{"profile_id"}).AddRow(tt.returnedID))

			id, err := repo.StoreProfile(tt.inputProfile)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.returnedID, id)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSQL_StoreUser(t *testing.T) {
	tests := []struct {
		name      string
		user      model.User
		returnID  int
		expectErr bool
	}{
		{
			name: "successfully inserts user",
			user: model.User{
				Login:    "testuser",
				Email:    "test@example.com",
				Password: "hashedpass",
				Phone:    "+1234567890",
				Status:   1,
				UserId:   1,
			},
			returnID:  1,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			conn, err := db.Conn(context.Background())
			require.NoError(t, err)
			defer conn.Close()

			repo := &repository.UserRepo{DB: conn}

			mock.ExpectQuery(regexp.QuoteMeta(repository.CreateUserQuery)).
				WithArgs(tt.user.Login, tt.user.Email, tt.user.Phone, tt.user.Password, tt.user.Status, tt.user.UserId).
				WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(tt.returnID))

			id, err := repo.StoreUser(tt.user)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.returnID, id)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSQL_CreateSession(t *testing.T) {
	tests := []struct {
		name      string
		sessionID int
		data      string
		ttl       time.Duration
		expectErr bool
	}{
		{
			name:      "successfully creates session",
			sessionID: 123,
			data:      "user:1",
			ttl:       10 * time.Minute,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			conn, err := db.Conn(context.Background())
			require.NoError(t, err)
			defer conn.Close()

			repo := &repository.UserRepo{DB: conn}

			mock.ExpectQuery(regexp.QuoteMeta(repository.StoreSessionQuery)).
				WithArgs(tt.sessionID, tt.data).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

			err = repo.StoreSession(tt.sessionID, tt.data)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSQL_SetLike(t *testing.T) {
	tests := []struct {
		name      string
		from      int
		to        int
		status    int
		likeID    int
		expectErr bool
	}{
		{
			name:      "successfully inserts like",
			from:      1,
			to:        2,
			status:    1,
			likeID:    42,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			conn, err := db.Conn(context.Background())
			require.NoError(t, err)
			defer conn.Close()

			repo := &repository.UserRepo{DB: conn}

			mock.ExpectQuery(regexp.QuoteMeta(repository.CreateLikeQuery)).
				WithArgs(tt.from, tt.to, tt.status).
				WillReturnRows(sqlmock.NewRows([]string{"like_id"}).AddRow(tt.likeID))

			likeID, err := repo.SetLike(tt.from, tt.to, tt.status)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.likeID, likeID)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSQL_StorePhoto(t *testing.T) {
	tests := []struct {
		name      string
		userID    int
		url       string
		expectErr bool
	}{
		{
			name:      "successfully stores photo",
			userID:    1,
			url:       "https://example.com/photo.jpg",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			conn, err := db.Conn(context.Background())
			require.NoError(t, err)
			defer conn.Close()

			repo := &repository.UserRepo{DB: conn}

			mock.ExpectExec(regexp.QuoteMeta(repository.UploadPhotoQuery)).
				WithArgs(tt.userID, tt.url).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err = repo.StorePhoto(tt.userID, tt.url)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSQL_GetPhotos(t *testing.T) {
	tests := []struct {
		name       string
		userID     int
		photoPaths []string
		expectErr  bool
	}{
		{
			name:       "successfully retrieves photo paths",
			userID:     1,
			photoPaths: []string{"/photo1.jpg", "/photo2.jpg"},
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			conn, err := db.Conn(context.Background())
			require.NoError(t, err)
			defer conn.Close()

			repo := &repository.UserRepo{DB: conn}

			rows := sqlmock.NewRows([]string{"path"})
			for _, path := range tt.photoPaths {
				rows.AddRow(path)
			}

			mock.ExpectQuery(regexp.QuoteMeta(repository.GetPhotoPathsQuery)).
				WithArgs(tt.userID).
				WillReturnRows(rows)

			photos, err := repo.GetPhotos(tt.userID)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.photoPaths, photos)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSQL_DeleteUserById(t *testing.T) {
	tests := []struct {
		name          string
		userID        int
		profileID     int
		queryRowErr   error
		deleteProfile error
		deleteUser    error
		expectedErr   error
	}{
		{
			name:        "successfully deletes user and profile",
			userID:      1,
			profileID:   10,
			expectedErr: nil,
		},
		{
			name:          "returns ErrDeleteProfile on profile delete error",
			userID:        3,
			profileID:     30,
			deleteProfile: errors.New("delete profile failed"),
			expectedErr:   model.ErrDeleteProfile,
		},
		{
			name:        "returns ErrDeleteUser on user delete error",
			userID:      4,
			profileID:   40,
			deleteUser:  errors.New("delete user failed"),
			expectedErr: model.ErrDeleteUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			conn, err := db.Conn(context.Background())
			require.NoError(t, err)
			defer conn.Close()

			repo := &repository.UserRepo{DB: conn}

			if tt.queryRowErr != nil {
				mock.ExpectQuery(regexp.QuoteMeta(repository.FindUserProfileQuery)).
					WithArgs(tt.userID).
					WillReturnError(tt.queryRowErr)
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(repository.FindUserProfileQuery)).
					WithArgs(tt.userID).
					WillReturnRows(sqlmock.NewRows([]string{"profile_id"}).AddRow(tt.profileID))

				if tt.deleteProfile != nil {
					mock.ExpectExec(regexp.QuoteMeta(repository.DeleteProfileQuery)).
						WithArgs(tt.profileID).
						WillReturnError(tt.deleteProfile)
				} else {
					mock.ExpectExec(regexp.QuoteMeta(repository.DeleteProfileQuery)).
						WithArgs(tt.profileID).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}

				if tt.deleteProfile == nil {
					if tt.deleteUser != nil {
						mock.ExpectExec(regexp.QuoteMeta(repository.DeleteUserQuery)).
							WithArgs(tt.userID).
							WillReturnError(tt.deleteUser)
					} else {
						mock.ExpectExec(regexp.QuoteMeta(repository.DeleteUserQuery)).
							WithArgs(tt.userID).
							WillReturnResult(sqlmock.NewResult(1, 1))
					}
				}
			}

			err = repo.DeleteUserById(tt.userID)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
