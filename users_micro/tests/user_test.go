package repository

import (
	"testing"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserRepo_GetUserByLogin(t *testing.T) {
	mockDB := new(MockDB)
	repo := &repository.UserRepo{DB: mockDB}

	login := "testlogin"
	user := model.User{
		UserId:   1,
		Login:    login,
		Email:    "email@example.com",
		Password: "pass",
		Phone:    "123456",
		Status:   1,
	}

	mockDB.On("QueryRow", mock.Anything, repository.GetUserByLoginQuery, []interface{}{login}).
		Return(&MockRow{
			data: []interface{}{
				user.UserId,
				user.Login,
				user.Email,
				user.Password,
				user.Phone,
				user.Status,
			},
			err: nil,
		})

	gotUser, err := repo.GetUserByLogin(login)

	assert.NoError(t, err)
	assert.Equal(t, user.UserId, gotUser.UserId)
	assert.Equal(t, user.Login, gotUser.Login)
	assert.Equal(t, user.Email, gotUser.Email)
	assert.Equal(t, user.Password, gotUser.Password)
	assert.Equal(t, user.Phone, gotUser.Phone)
	assert.Equal(t, user.Status, gotUser.Status)

	mockDB.AssertExpectations(t)
}

func TestUserRepo_StoreUser(t *testing.T) {
	mockDB := new(MockDB)
	repo := &repository.UserRepo{DB: mockDB}

	user := model.User{
		Login:    "login",
		Email:    "email@mail.com",
		Phone:    "123",
		Password: "password",
		Status:   1,
		UserId:   42,
	}

	mockDB.On("QueryRow", mock.Anything, repository.CreateUserQuery,
		[]interface{}{user.Login, user.Email, user.Phone, user.Password, user.Status, user.UserId}).
		Return(&MockRow{
			data: []interface{}{100},
			err:  nil,
		})

	userId, err := repo.StoreUser(user)

	assert.NoError(t, err)
	assert.Equal(t, 100, userId)

	mockDB.AssertExpectations(t)
}

func TestUserRepo_DeleteUserById(t *testing.T) {
	mockDB := new(MockDB)
	repo := &repository.UserRepo{DB: mockDB}

	userId := 123

	mockDB.On("Exec", mock.Anything, repository.DeleteUserQuery, []interface{}{userId}).
		Return(nil, nil)

	err := repo.DeleteUserById(userId)
	assert.NoError(t, err)

	mockDB.AssertExpectations(t)
}

func TestUserRepo_UserExists(t *testing.T) {
	mockDB := new(MockDB)
	repo := &repository.UserRepo{DB: mockDB}

	login := "userexists"

	mockDB.On("QueryRow", mock.Anything, repository.GetUserByLoginQuery, []interface{}{login}).
		Return(&MockRow{
			data: []interface{}{
				1, login, "email", "pass", "phone", 1,
			},
			err: nil,
		})

	exists := repo.UserExists(login)
	assert.True(t, exists)

	mockDB.AssertExpectations(t)
}

func TestUserRepo_StoreSession(t *testing.T) {
	mockDB := new(MockDB)
	repo := &repository.UserRepo{DB: mockDB}

	userID := 5
	sessionID := "sess123"

	mockDB.On("QueryRow", mock.Anything, repository.StoreSessionQuery,
		[]interface{}{userID, sessionID}).
		Return(&MockRow{
			data: []interface{}{10},
			err:  nil,
		})

	err := repo.StoreSession(userID, sessionID)
	assert.NoError(t, err)

	mockDB.AssertExpectations(t)
}

func TestUserRepo_DeleteSession(t *testing.T) {
	mockDB := new(MockDB)
	repo := &repository.UserRepo{DB: mockDB}

	userId := 7
	sessionId := 15

	mockDB.On("QueryRow", mock.Anything, repository.FindSessionQuery, []interface{}{userId}).
		Return(&MockRow{
			data: []interface{}{sessionId},
			err:  nil,
		})

	mockDB.On("Exec", mock.Anything, repository.DeleteSessionQuery, []interface{}{userId}).
		Return(nil, nil)

	err := repo.DeleteSession(userId)
	assert.NoError(t, err)

	mockDB.AssertExpectations(t)
}

func TestUserRepo_GetUserParams(t *testing.T) {
	mockDB := new(MockDB)
	repo := &repository.UserRepo{DB: mockDB}

	userID := 33
	user := model.User{
		Login:  "login33",
		Email:  "email33",
		Phone:  "phone33",
		Status: 1,
	}

	mockDB.On("QueryRow", mock.Anything, repository.GetUserByIdQuery, []interface{}{userID}).
		Return(&MockRow{
			data: []interface{}{
				user.Login,
				user.Email,
				user.Phone,
				user.Status,
			},
			err: nil,
		})

	gotUser, err := repo.GetUserParams(userID)

	assert.NoError(t, err)
	assert.Equal(t, user.Login, gotUser.Login)
	assert.Equal(t, user.Email, gotUser.Email)
	assert.Equal(t, user.Phone, gotUser.Phone)
	assert.Equal(t, user.Status, gotUser.Status)

	mockDB.AssertExpectations(t)
}

func TestUserRepo_GetAdmin(t *testing.T) {
	mockDB := new(MockDB)
	repo := &repository.UserRepo{DB: mockDB}

	userID := 8
	exists := true

	mockDB.On("QueryRow", mock.Anything, repository.GetAdminQuery, []interface{}{userID}).
		Return(&MockRow{
			data: []interface{}{exists},
			err:  nil,
		})

	gotExists, err := repo.GetAdmin(userID)

	assert.NoError(t, err)
	assert.Equal(t, exists, gotExists)

	mockDB.AssertExpectations(t)
}
