package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-redis/redis/v8"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type ChatRepository interface {
	GetChats(userID int) ([]model.Chat, error)
	GetChatParticipants(chatID int) (int, int, error)
	CreateChat(firstProfileID, secondProfileID int) (int, error)
	DeleteChat(firstID int, secondID int) error

	GetMessages(chatID int) ([]model.Message, error)
	DeleteMessage(messageID int, chatID int) error
	CreateMessage(chatID int, userID int, content string, status int) (int, error)
	GetMessagesFromCache(chatID int) ([]model.Message, error)
	UpdateMessageStatus(chatID int) error
}

type ChatRepo struct {
	DB     *sql.DB
	client *redis.Client
	ctx    context.Context
}

func NewChatRepo() (*ChatRepo, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return &ChatRepo{}, err
	}

	cfg := InitPostgresConfig()
	db, err := InitPostgresConnection(cfg)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return &ChatRepo{}, err
	}

	return &ChatRepo{
		DB:     db,
		client: client,
		ctx:    ctx,
	}, nil
}

func (cr *ChatRepo) CloseRepo() error {
	return cr.client.Close()
}

const GetChatParticipantsQuery = `
		SELECT 
			first_profile_id,
			second_profile_id 
		FROM chats WHERE chat_id = $1;`

func (cr *ChatRepo) GetChatParticipants(chatID int) (int, int, error) {
	var firstID, secondID int
	err := cr.DB.QueryRowContext(context.Background(),
		GetChatParticipantsQuery,
		chatID,
	).Scan(&firstID, &secondID)
	if err != nil {
		return 0, 0, err
	}
	return firstID, secondID, nil
}

const (
	GetChatsQuery = `
	SELECT 
		chat_id, 
		first_profile_id, 
		second_profile_id, 
		last_message 
	FROM chats 
	WHERE (first_profile_id = $1 OR second_profile_id = $1);
	`
	GetProfileParams = `
	SELECT 
		p.firstname, 
		p.lastname, 
		p.description, 
		s.path AS avatar
	FROM profiles p
	LEFT JOIN "static" s 
	ON p.profile_id = s.profile_id 
	WHERE p.profile_id = $1;
	`
)

func (cr *ChatRepo) GetChats(userID int) ([]model.Chat, error) {
	rows, err := cr.DB.QueryContext(context.Background(), GetChatsQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []model.Chat
	for rows.Next() {
		var chat model.Chat
		var firstID, secondID int
		if err := rows.Scan(&chat.ChatId, &firstID, &secondID, &chat.LastMessage); err != nil {
			return nil, err
		}

		var reqID int
		if firstID == userID {
			reqID = secondID
		} else {
			reqID = firstID
		}

		var firstName, lastName, description, avatar string
		err := cr.DB.QueryRowContext(context.Background(), GetProfileParams, reqID).Scan(
			&firstName,
			&lastName,
			&description,
			&avatar,
		)
		if err != nil {
			return nil, err
		}

		chat.ProfileId = reqID
		chat.ProfileDescription = description
		chat.ProfilePicture = avatar
		chat.ProfileName = firstName + " " + lastName

		redisKey := fmt.Sprintf("chat:%d:messages", chat.ChatId)

		_, err = cr.client.Get(cr.ctx, redisKey).Result()
		if err != nil {
			if err != redis.Nil {
				return nil, err
			}
			chat.IsRead = false
		} else {
			chat.IsRead = true
		}
		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}

const CreateChatQuery = `
		INSERT INTO chats (first_profile_id, second_profile_id, last_message)
		VALUES ($1, $2, $3)
		ON CONFLICT (first_profile_id, second_profile_id) DO NOTHING
		RETURNING chat_id;
	`

func (cr *ChatRepo) CreateChat(firstProfileID, secondProfileID int) (int, error) {
	var chatID int
	if firstProfileID > secondProfileID {
		firstProfileID, secondProfileID = secondProfileID, firstProfileID
	}
	err := cr.DB.QueryRowContext(context.Background(), CreateChatQuery, firstProfileID, secondProfileID, "").Scan(&chatID)
	if err != nil {
		return 0, err
	}

	return chatID, nil
}

const DeleteChatBetweenUsersQuery = `
	DELETE FROM chats
	WHERE (first_profile_id = $1 AND second_profile_id = $2)
	   OR (first_profile_id = $2 AND second_profile_id = $1);
`

func (cr *ChatRepo) DeleteChat(firstID int, secondID int) error {
	_, err := cr.DB.ExecContext(context.Background(), DeleteChatBetweenUsersQuery, firstID, secondID)
	return err
}

const GetMessagesQuery = `
	SELECT 
		message_id,
		user_id,
		content,
		status 
	FROM messages
	WHERE chat_id = $1; 
`

func (cr *ChatRepo) GetMessages(chatID int) ([]model.Message, error) {
	rows, err := cr.DB.QueryContext(context.Background(), GetMessagesQuery, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var message model.Message
		if err := rows.Scan(&message.MessageID, &message.SenderID, &message.Text, &message.Status); err != nil {
			return nil, err
		}

		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

const (
	DeleteMessageQuery = `
		DELETE FROM messages
		WHERE chat_id = $1 AND message_id = $2;
	`

	GetLastMessageQuery = `
		SELECT content FROM messages
		WHERE chat_id = $1
		ORDER BY created_at DESC
		LIMIT 1;
	`

	UpdateLastMessageQuery = `
		UPDATE chats
		SET last_message = $1
		WHERE chat_id = $2;
	`
)

func (cr *ChatRepo) DeleteMessage(messageID int, chatID int) error {
	tx, err := cr.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(context.Background(), DeleteMessageQuery, chatID, messageID)
	if err != nil {
		return err
	}

	var lastMsg string
	err = tx.QueryRowContext(context.Background(), GetLastMessageQuery, chatID).Scan(&lastMsg)

	if err == sql.ErrNoRows {
		lastMsg = ""
	} else if err != nil {
		return err
	}

	_, err = tx.ExecContext(context.Background(), UpdateLastMessageQuery, lastMsg, chatID)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	redisKey := fmt.Sprintf("chat:%d:messages", chatID)
	existingMessages, err := cr.GetMessagesFromCache(chatID)
	if err != nil {
		return err
	}

	var updatedMessages []model.Message
	for _, msg := range existingMessages {
		if msg.MessageID != messageID {
			updatedMessages = append(updatedMessages, msg)
		}
	}
	messageJSON, err := json.Marshal(updatedMessages)
	if err != nil {
		return err
	}

	_, err = cr.client.Set(cr.ctx, redisKey, messageJSON, 0).Result()
	return err
}

const (
	InsertMessageQuery = `
		INSERT INTO messages (chat_id, user_id, content, status)
		VALUES ($1, $2, $3, $4)
		RETURNING message_id;
	`

	UpdateChatLastMessageQuery = `
		UPDATE chats
		SET last_message = $1
		WHERE chat_id = $2;
	`
)

func (cr *ChatRepo) CreateMessage(chatID int, userID int, content string, status int) (int, error) {
	tx, err := cr.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var messageID int
	err = tx.QueryRowContext(context.Background(), InsertMessageQuery, chatID, userID, content, status).Scan(&messageID)
	if err != nil {
		return 0, err
	}

	_, err = tx.ExecContext(context.Background(), UpdateChatLastMessageQuery, content, chatID)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	redisKey := fmt.Sprintf("chat:%d:messages", chatID)

	message := model.Message{
		MessageID: messageID,
		SenderID:  userID,
		Text:      content,
		Status:    status,
	}

	existingMessages, err := cr.GetMessagesFromCache(chatID)
	if err != nil {
		return 0, err
	}

	existingMessages = append(existingMessages, message)

	messageJSON, err := json.Marshal(existingMessages)
	if err != nil {
		return 0, err
	}

	_, err = cr.client.Set(cr.ctx, redisKey, messageJSON, 0).Result()
	if err != nil {
		return 0, err
	}

	cr.client.LTrim(cr.ctx, redisKey, 0, 49)

	return messageID, nil
}

const (
	UpdateMessageStatusQuery = `
		UPDATE messages
		SET status = 2
		WHERE chat_id = $1;
	`
)

func (cr *ChatRepo) UpdateMessageStatus(chatID int) error {
	redisKey := fmt.Sprintf("chat:%d:messages", chatID)

	_, err := cr.DB.ExecContext(context.Background(), UpdateMessageStatusQuery, chatID)
	if err != nil {
		return err
	}
	_, err = cr.client.Del(cr.ctx, redisKey).Result()
	return err
}

func (cr *ChatRepo) GetMessagesFromCache(chatID int) ([]model.Message, error) {
	redisKey := fmt.Sprintf("chat:%d:messages", chatID)

	result, err := cr.client.Get(cr.ctx, redisKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var messages []model.Message
	err = json.Unmarshal([]byte(result), &messages)
	if err != nil {
		return nil, err
	}

	return messages, nil
}
