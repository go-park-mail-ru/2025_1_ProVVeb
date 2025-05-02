package tests

import (
	"context"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/mocks"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGetUserByLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)

	expected := model.User{
		UserId:   1,
		Login:    "testuser",
		Email:    "test@example.com",
		Password: "hashed_password",
		Phone:    "+1234567890",
		Status:   1,
	}

	mockRepo.EXPECT().
		GetUserByLogin(gomock.Any(), "testuser").
		Return(expected, nil)

	user, err := mockRepo.GetUserByLogin(context.Background(), "testuser")

	require.NoError(t, err)
	require.Equal(t, expected, user)
}
func TestSessionRepositoryMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSessionRepository(ctrl)

	type testCase struct {
		name      string
		callMock  func()
		runTest   func() error
		expectErr bool
	}

	sessionID := "session123"
	data := "user:1"
	ttl := 10 * time.Minute

	tests := []testCase{
		{
			name: "GetSession success",
			callMock: func() {
				mockRepo.EXPECT().
					GetSession(sessionID).
					Return(data, nil)
			},
			runTest: func() error {
				res, err := mockRepo.GetSession(sessionID)
				require.Equal(t, data, res)
				return err
			},
		},
		{
			name: "DeleteSession success",
			callMock: func() {
				mockRepo.EXPECT().
					DeleteSession(sessionID).
					Return(nil)
			},
			runTest: func() error {
				return mockRepo.DeleteSession(sessionID)
			},
		},
		{
			name: "StoreSession success",
			callMock: func() {
				mockRepo.EXPECT().
					StoreSession(sessionID, data, ttl).
					Return(nil)
			},
			runTest: func() error {
				return mockRepo.StoreSession(sessionID, data, ttl)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.callMock()
			err := tc.runTest()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUserRepositoryMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)

	type testCase struct {
		name      string
		callMock  func()
		runTest   func() error
		expectErr bool
	}

	user := model.User{
		Login:    "testuser",
		Email:    "user@example.com",
		Password: "hashedpass",
		Phone:    "+1234567890",
		Status:   1,
	}
	expectedUserID := 1

	profile := model.Profile{
		FirstName:   "Алиса",
		LastName:    "Сидорова",
		IsMale:      false,
		Birthday:    time.Date(1995, 5, 5, 0, 0, 0, 0, time.UTC),
		Height:      165,
		Description: "Описание профиля",
	}
	expectedProfileID := 42

	expectedMatches := []model.Profile{
		{
			ProfileId:   2,
			FirstName:   "Иван",
			LastName:    "Иванов",
			IsMale:      true,
			Birthday:    time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			Height:      180,
			Description: "Описание",
		},
	}

	tests := []testCase{
		{
			name: "StoreUser success",
			callMock: func() {
				mockRepo.EXPECT().
					StoreUser(user).
					Return(expectedUserID, nil)
			},
			runTest: func() error {
				id, err := mockRepo.StoreUser(user)
				require.Equal(t, expectedUserID, id)
				return err
			},
		},
		{
			name: "StoreProfile success",
			callMock: func() {
				mockRepo.EXPECT().
					StoreProfile(profile).
					Return(expectedProfileID, nil)
			},
			runTest: func() error {
				id, err := mockRepo.StoreProfile(profile)
				require.Equal(t, expectedProfileID, id)
				return err
			},
		},
		{
			name: "GetMatches success",
			callMock: func() {
				mockRepo.EXPECT().
					GetMatches(1).
					Return(expectedMatches, nil)
			},
			runTest: func() error {
				res, err := mockRepo.GetMatches(1)
				require.Equal(t, expectedMatches, res)
				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.callMock()
			err := tc.runTest()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSupportInterfaces(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStatic := mocks.NewMockStaticRepository(ctrl)
	mockHasher := mocks.NewMockPasswordHasher(ctrl)
	mockValidator := mocks.NewMockUserParamsValidator(ctrl)

	type testCase struct {
		name      string
		callMock  func()
		runTest   func() error
		expectErr bool
	}

	urls := []string{"img1.jpg", "img2.jpg"}
	expectedImages := [][]byte{[]byte("data1"), []byte("data2")}

	password := "supersecret"
	hashed := "hashedsecret"

	login := "validlogin"
	pass := "StrongPass123"

	tests := []testCase{
		{
			name: "StaticRepository GetImages",
			callMock: func() {
				mockStatic.EXPECT().
					GetImages(urls).
					Return(expectedImages, nil)
			},
			runTest: func() error {
				data, err := mockStatic.GetImages(urls)
				require.Equal(t, expectedImages, data)
				return err
			},
		},
		{
			name: "StaticRepository UploadImages",
			callMock: func() {
				mockStatic.EXPECT().
					UploadImages([]byte("imgdata"), "avatar.jpg", "image/jpeg").
					Return(nil)
			},
			runTest: func() error {
				return mockStatic.UploadImages([]byte("imgdata"), "avatar.jpg", "image/jpeg")
			},
		},
		{
			name: "PasswordHasher Hash",
			callMock: func() {
				mockHasher.EXPECT().
					Hash(password).
					Return(hashed)
			},
			runTest: func() error {
				res := mockHasher.Hash(password)
				require.Equal(t, hashed, res)
				return nil
			},
		},
		{
			name: "PasswordHasher Compare",
			callMock: func() {
				mockHasher.EXPECT().
					Compare(hashed, password).
					Return(true)
			},
			runTest: func() error {
				ok := mockHasher.Compare(hashed, password)
				require.True(t, ok)
				return nil
			},
		},
		{
			name: "UserParamsValidator ValidateLogin",
			callMock: func() {
				mockValidator.EXPECT().
					ValidateLogin(login).
					Return(nil)
			},
			runTest: func() error {
				return mockValidator.ValidateLogin(login)
			},
		},
		{
			name: "UserParamsValidator ValidatePassword",
			callMock: func() {
				mockValidator.EXPECT().
					ValidatePassword(pass).
					Return(nil)
			},
			runTest: func() error {
				return mockValidator.ValidatePassword(pass)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.callMock()
			err := tc.runTest()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSupportUserRepositoryMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)

	type testCase struct {
		name      string
		callMock  func()
		runTest   func() error
		expectErr bool
	}

	expectedProfile := model.Profile{
		ProfileId:   1,
		FirstName:   "Анна",
		LastName:    "Петрова",
		IsMale:      false,
		Birthday:    time.Date(1993, 3, 3, 0, 0, 0, 0, time.UTC),
		Height:      170,
		Description: "Тестовый профиль",
	}

	expectedProfiles := []model.Profile{
		{
			ProfileId:   1,
			FirstName:   "Анна",
			LastName:    "Петрова",
			IsMale:      false,
			Birthday:    time.Date(1993, 3, 3, 0, 0, 0, 0, time.UTC),
			Height:      170,
			Description: "Тестовый профиль",
		},
	}

	expectedLikeID := 101
	expectedPhotoPaths := []string{"photo1.jpg", "photo2.jpg"}

	tests := []testCase{
		{
			name: "DeleteSession success",
			callMock: func() {
				mockRepo.EXPECT().
					DeleteSession(1).
					Return(nil)
			},
			runTest: func() error {
				return mockRepo.DeleteSession(1)
			},
		},
		{
			name: "DeleteUserById success",
			callMock: func() {
				mockRepo.EXPECT().
					DeleteUserById(1).
					Return(nil)
			},
			runTest: func() error {
				return mockRepo.DeleteUserById(1)
			},
		},
		{
			name: "GetProfileById success",
			callMock: func() {
				mockRepo.EXPECT().
					GetProfileById(1).
					Return(expectedProfile, nil)
			},
			runTest: func() error {
				profile, err := mockRepo.GetProfileById(1)
				require.Equal(t, expectedProfile, profile)
				return err
			},
		},
		{
			name: "GetProfilesByUserId success",
			callMock: func() {
				mockRepo.EXPECT().
					GetProfilesByUserId(1).
					Return(expectedProfiles, nil)
			},
			runTest: func() error {
				profiles, err := mockRepo.GetProfilesByUserId(1)
				require.Equal(t, expectedProfiles, profiles)
				return err
			},
		},
		{
			name: "SetLike success",
			callMock: func() {
				mockRepo.EXPECT().
					SetLike(1, 2, 1).
					Return(expectedLikeID, nil)
			},
			runTest: func() error {
				likeID, err := mockRepo.SetLike(1, 2, 1)
				require.Equal(t, expectedLikeID, likeID)
				return err
			},
		},
		{
			name: "StorePhoto success",
			callMock: func() {
				mockRepo.EXPECT().
					StorePhoto(1, "photo.jpg").
					Return(nil)
			},
			runTest: func() error {
				return mockRepo.StorePhoto(1, "photo.jpg")
			},
		},
		{
			name: "GetPhotos success",
			callMock: func() {
				mockRepo.EXPECT().
					GetPhotos(1).
					Return(expectedPhotoPaths, nil)
			},
			runTest: func() error {
				paths, err := mockRepo.GetPhotos(1)
				require.Equal(t, expectedPhotoPaths, paths)
				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.callMock()
			err := tc.runTest()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
