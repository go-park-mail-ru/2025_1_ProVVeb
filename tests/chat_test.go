package tests

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func newTestChatRepo(t *testing.T) (*repository.ChatRepo, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisServer.Addr(),
	})

	repo := &repository.ChatRepo{
		DB:     db,
		Client: redisClient,
		Ctx:    context.Background(),
	}

	cleanup := func() {
		db.Close()
		redisClient.Close()
		redisServer.Close()
	}

	return repo, mock, cleanup
}

func TestChatRepo_GetChats(t *testing.T) {
	db, mock, _ := sqlmock.New()
	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("could not start miniredis server: %v", err)
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisServer.Addr(),
	})
	repo := &repository.ChatRepo{
		DB:     db,
		Client: redisClient,
		Ctx:    context.Background(),
	}

	userID := 1

	rows := sqlmock.NewRows([]string{"chat_id", "first_profile_id", "second_profile_id", "last_message", "last_sender"}).
		AddRow(1, 1, 2, "Hello", 1).
		AddRow(2, 1, 3, "How are you?", 2)
	mock.ExpectQuery("SELECT chat_id, first_profile_id, second_profile_id, last_message, last_sender FROM chats WHERE").
		WithArgs(userID).
		WillReturnRows(rows)

	redisServer.Set("chat:1:messages_user1", "")
	redisServer.Set("chat:2:messages_user1", "")

	_, err = repo.GetChats(userID)

}

func TestChatRepo_CreateChat(t *testing.T) {
	repo, mock, cleanup := newTestChatRepo(t)
	defer cleanup()

	firstProfileID := 1
	secondProfileID := 2
	expectedChatID := 3

	mock.ExpectQuery("INSERT INTO chats").
		WithArgs(firstProfileID, secondProfileID, "", secondProfileID).
		WillReturnRows(sqlmock.NewRows([]string{"chat_id"}).AddRow(expectedChatID))

	chatID, err := repo.CreateChat(firstProfileID, secondProfileID)
	assert.NoError(t, err)
	assert.Equal(t, expectedChatID, chatID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestChatRepo_UpdateMessageStatus(t *testing.T) {
	repo, mock, cleanup := newTestChatRepo(t)
	defer cleanup()

	chatID := 1
	userID := 1

	mock.ExpectExec("UPDATE messages SET status = 2 WHERE chat_id = \\$1 AND user_id = \\$2").
		WithArgs(chatID, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery(`SELECT first_profile_id, second_profile_id FROM chats WHERE chat_id = \$1`).
		WithArgs(chatID).
		WillReturnRows(sqlmock.NewRows([]string{"first_profile_id", "second_profile_id"}).
			AddRow(1, 2))

	repo.Client.Set(repo.Ctx, "chat:1:messages_user1", "[]", 0)

	err := repo.UpdateMessageStatus(chatID, userID)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}
