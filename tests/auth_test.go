package tests

import (
	"context"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	auth "github.com/go-park-mail-ru/2025_1_ProVVeb/auth_micro/server"
	"github.com/go-redis/redis/v8"

	model "github.com/go-park-mail-ru/2025_1_ProVVeb/auth_micro/config"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSessionRepo_CreateAndStoreSession(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := &auth.SessionRepo{
		DB:     db,
		Client: rdb,
		Ctx:    context.Background(),
	}

	session := repo.CreateSession(123)
	testData := "some_data"

	mock.ExpectQuery(regexp.QuoteMeta(`
	INSERT INTO sessions (user_id, token, created_at, expires_at)
	VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + INTERVAL '72 hours')
	RETURNING id;
	`)).WithArgs(123, testData).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := repo.StoreSession(123, session.SessionId, testData, session.Expires)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
	val, _ := mr.Get(session.SessionId)
	assert.Equal(t, strconv.Itoa(session.UserId), val)
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

	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo.DB = db
	sessionID := "test-session"
	data := "user data"
	ttl := 10 * time.Second

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO sessions (user_id, token, created_at, expires_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + INTERVAL '72 hours')
		RETURNING id;
	`)).WithArgs(123, data).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := repo.StoreSession(123, sessionID, data, ttl)
	assert.NoError(t, err)

	val, err := repo.GetSession(sessionID)
	assert.NoError(t, err)
	assert.Equal(t, "123", val)
}

func TestDeleteAllSessions(t *testing.T) {
	repo := initTestRepo(t)

	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo.DB = db
	userId1 := 123
	userId2 := 124
	sess1 := "sess1"
	data1 := "data1"
	sess2 := "sess2"
	data2 := "data2"

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO sessions (user_id, token, created_at, expires_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + INTERVAL '72 hours')
		RETURNING id;
	`)).WithArgs(userId1, data1).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO sessions (user_id, token, created_at, expires_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + INTERVAL '72 hours')
		RETURNING id;
	`)).WithArgs(userId2, data2).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	_ = repo.StoreSession(userId1, sess1, data1, 10*time.Second)
	_ = repo.StoreSession(userId2, sess2, data2, 10*time.Second)

	err := repo.DeleteAllSessions()
	assert.NoError(t, err)

	_, err = repo.GetSession(sess1)
	assert.Equal(t, model.ErrSessionNotFound, err)
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

	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo.DB = db
	sessionID := "sess123"
	data := "hello"

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO sessions (user_id, token, created_at, expires_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + INTERVAL '72 hours')
		RETURNING id;
	`)).WithArgs(123, data).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := repo.StoreSession(123, sessionID, data, 5*time.Second)
	assert.NoError(t, err)

	val, err := repo.GetSession(sessionID)
	assert.NoError(t, err)
	assert.Equal(t, "123", val)
}

func TestRemoveSessionEntry(t *testing.T) {
	repo := initTestRepo(t)

	sessionID := "delete_me"
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo.DB = db

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO sessions (user_id, token, created_at, expires_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + INTERVAL '72 hours')
		RETURNING id;
	`)).WithArgs(123, "bye").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id FROM sessions WHERE user_id = $1;
	`)).WithArgs(123).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM sessions WHERE user_id = $1;
	`)).WithArgs(123).WillReturnResult(sqlmock.NewResult(0, 1))

	_ = repo.StoreSession(123, sessionID, "bye", 5*time.Second)

	err := repo.Client.Set(repo.Ctx, sessionID, "123", 0).Err()
	assert.NoError(t, err)

	err = repo.DeleteSession("123")
	assert.NoError(t, err)

	_, err = repo.GetSession("123")
	assert.Equal(t, model.ErrSessionNotFound, err)
}

func TestAttemptCheckerLogic(t *testing.T) {
	repo := initTestRepo(t)
	ip := "192.168.1.1"

	blockUntil, err := repo.CheckAttempts(ip)
	assert.NoError(t, err)
	assert.Empty(t, blockUntil)

	for range model.MaxAttempts {
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
