package tests

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	auth "github.com/go-park-mail-ru/2025_1_ProVVeb/auth_micro/server"
	"github.com/go-redis/redis/v8"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/mocks"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSessionRepo_CreateAndStoreSession(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	ctx := context.Background()
	repo := &auth.SessionRepo{
		Client: rdb,
		Ctx:    ctx,
	}

	session := repo.CreateSession(123)

	assert.NotEmpty(t, session.SessionId)
	assert.Equal(t, 123, session.UserId)
	assert.Equal(t, model.SessionDuration, session.Expires)

	err = repo.StoreSession(session.SessionId, "some_data", session.Expires)
	assert.NoError(t, err)

	val, err := mr.Get(session.SessionId)
	assert.NoError(t, err)
	assert.Equal(t, "some_data", val)
}

func TestCreateSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSessionRepository(ctrl)

	userId := 123
	sessionId := "random-session-id"
	expectedSession := model.Session{
		SessionId: sessionId,
		UserId:    userId,
		Expires:   model.SessionDuration,
	}

	mockRepo.EXPECT().
		CreateSession(userId).
		Return(expectedSession).
		Times(1)

	session := mockRepo.CreateSession(userId)

	assert.Equal(t, expectedSession, session)
}

func TestGetSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSessionRepository(ctrl)

	sessionId := "random-session-id"
	expectedData := "session-data"

	mockRepo.EXPECT().
		GetSession(sessionId).
		Return(expectedData, nil).
		Times(1)

	data, err := mockRepo.GetSession(sessionId)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)
}

func TestDeleteSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSessionRepository(ctrl)

	sessionId := "random-session-id"

	mockRepo.EXPECT().
		DeleteSession(sessionId).
		Return(nil).
		Times(1)

	err := mockRepo.DeleteSession(sessionId)

	assert.NoError(t, err)
}

func TestCheckAttempts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSessionRepository(ctrl)

	ip := "192.168.1.1"
	blockUntil := "1622119871"

	mockRepo.EXPECT().
		CheckAttempts(ip).
		Return(blockUntil, nil).
		Times(1)

	blockTime, err := mockRepo.CheckAttempts(ip)

	assert.NoError(t, err)
	assert.Equal(t, blockUntil, blockTime)
}

func TestIncreaseAttempts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSessionRepository(ctrl)

	ip := "192.168.1.1"

	mockRepo.EXPECT().
		IncreaseAttempts(ip).
		Return(nil).
		Times(1)

	err := mockRepo.IncreaseAttempts(ip)

	assert.NoError(t, err)
}

func TestDeleteAttempts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSessionRepository(ctrl)

	ip := "192.168.1.1"

	mockRepo.EXPECT().
		DeleteAttempts(ip).
		Return(nil).
		Times(1)

	err := mockRepo.DeleteAttempts(ip)

	assert.NoError(t, err)
}

func TestStoreSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSessionRepository(ctrl)

	sessionId := "random-session-id"
	data := "session_data"
	ttl := time.Duration(12 * time.Hour)

	mockRepo.EXPECT().
		StoreSession(sessionId, data, ttl).
		Return(nil).
		Times(1)

	err := mockRepo.StoreSession(sessionId, data, ttl)

	assert.NoError(t, err)
}

func initTestRepo(t *testing.T) *auth.SessionRepo {
	mr, err := miniredis.Run()
	assert.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	repo := &auth.SessionRepo{
		Client: client,
		Ctx:    client.Context(),
	}
	return repo
}

func TestStoreSessionA(t *testing.T) {
	repo := initTestRepo(t)

	sessionID := "test-session"
	data := "user data"
	ttl := 10 * time.Second

	err := repo.StoreSession(sessionID, data, ttl)
	assert.NoError(t, err)

	val, err := repo.GetSession(sessionID)
	assert.NoError(t, err)
	assert.Equal(t, data, val)
}

func TestDeleteAllSessions(t *testing.T) {
	repo := initTestRepo(t)

	_ = repo.StoreSession("sess1", "data1", 10*time.Second)
	_ = repo.StoreSession("sess2", "data2", 10*time.Second)

	err := repo.DeleteAllSessions()
	assert.NoError(t, err)

	_, err = repo.GetSession("sess1")
	assert.Equal(t, model.ErrSessionNotFound, err)
}

func TestCloseRepo(t *testing.T) {
	repo := initTestRepo(t)

	err := repo.CloseRepo()
	assert.NoError(t, err)

	err = repo.StoreSession("test", "data", 1*time.Second)
	assert.Error(t, err)
}

func TestCreateSessionA(t *testing.T) {
	repo := initTestRepo(t)

	userId := 42
	session := repo.CreateSession(userId)

	assert.Equal(t, userId, session.UserId)
	assert.Len(t, session.SessionId, model.SessionIdLength)
	assert.Equal(t, model.SessionDuration, session.Expires)
}

func TestRetrieveSessionData(t *testing.T) {
	repo := initTestRepo(t)

	sessionID := "sess123"
	data := "hello"
	err := repo.StoreSession(sessionID, data, 5*time.Second)
	assert.NoError(t, err)

	val, err := repo.GetSession(sessionID)
	assert.NoError(t, err)
	assert.Equal(t, data, val)
}

func TestRemoveSessionEntry(t *testing.T) {
	repo := initTestRepo(t)

	sessionID := "delete_me"
	_ = repo.StoreSession(sessionID, "bye", 5*time.Second)

	err := repo.DeleteSession(sessionID)
	assert.NoError(t, err)

	_, err = repo.GetSession(sessionID)
	assert.Equal(t, model.ErrSessionNotFound, err)
}

func TestAttemptCheckerLogic(t *testing.T) {
	repo := initTestRepo(t)
	ip := "192.168.1.1"

	blockUntil, err := repo.CheckAttempts(ip)
	assert.NoError(t, err)
	assert.Empty(t, blockUntil)

	for i := 0; i < model.MaxAttempts; i++ {
		err := repo.IncreaseAttempts(ip)
		assert.NoError(t, err)
	}

	_, err = repo.CheckAttempts(ip)
	assert.Error(t, err)
}

func TestAddLoginAttempt(t *testing.T) {
	repo := initTestRepo(t)
	ip := "10.0.0.1"

	for i := 1; i <= model.MaxAttempts+1; i++ {
		err := repo.IncreaseAttempts(ip)
		assert.NoError(t, err)
	}
	blockKey := model.TimeAttemptsKeyPrefix + ip
	val, err := repo.Client.Get(repo.Ctx, blockKey).Result()
	assert.NoError(t, err)
	assert.NotEmpty(t, val)
}

func TestClearAttemptCounters(t *testing.T) {
	repo := initTestRepo(t)
	ip := "127.0.0.1"

	_ = repo.IncreaseAttempts(ip)
	_ = repo.DeleteAttempts(ip)

	countKey := model.AttemptsKeyPrefix + ip
	timeKey := model.TimeAttemptsKeyPrefix + ip

	_, err := repo.Client.Get(repo.Ctx, countKey).Result()
	assert.ErrorIs(t, err, redis.Nil)

	_, err = repo.Client.Get(repo.Ctx, timeKey).Result()
	assert.ErrorIs(t, err, redis.Nil)
}
