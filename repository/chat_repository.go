package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

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
	GetMessagesFromCache(chatID int, userID int) ([]model.Message, error)
	UpdateMessageStatus(chatID int, userID int) error

	updateMessageCache(chatID, userID int, messages []model.Message) error
}

type ChatRepo struct {
	DB     *sql.DB
	Client *redis.Client
	Ctx    context.Context
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
		Client: client,
		Ctx:    ctx,
	}, nil
}

func (cr *ChatRepo) CloseRepo() error {
	return cr.Client.Close()
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
	SELECT DISTINCT ON (c.chat_id) 
    c.chat_id, 
    c.first_profile_id, 
    c.second_profile_id, 
    c.last_message,
    c.last_sender 
FROM chats c
JOIN users u1 ON u1.profile_id = c.first_profile_id
JOIN users u2 ON u2.profile_id = c.second_profile_id
LEFT JOIN blacklist b1 ON b1.user_id = u1.user_id
LEFT JOIN blacklist b2 ON b2.user_id = u2.user_id
WHERE 
    (c.first_profile_id = $1 OR c.second_profile_id = $1)
    AND b1.user_id IS NULL
    AND b2.user_id IS NULL;
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
		var sender int
		if err := rows.Scan(&chat.ChatId, &firstID, &secondID, &chat.LastMessage, &sender); err != nil {
			return nil, err
		}

		chat.IsSelf = false
		if sender == userID {
			chat.IsSelf = true
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

		redisKey := fmt.Sprintf("chat:%d:messages_user%d", chat.ChatId, userID)

		q, err := cr.Client.Get(cr.Ctx, redisKey).Result()
		if err == redis.Nil || q == "" || q == "null" {
			chat.IsRead = false
		} else if err != nil {
			return nil, err
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
		INSERT INTO chats (first_profile_id, second_profile_id, last_message, last_sender)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (first_profile_id, second_profile_id) DO NOTHING
		RETURNING chat_id;
	`

func (cr *ChatRepo) CreateChat(firstProfileID, secondProfileID int) (int, error) {
	var chatID int
	if firstProfileID > secondProfileID {
		firstProfileID, secondProfileID = secondProfileID, firstProfileID
	}
	err := cr.DB.QueryRowContext(context.Background(),
		CreateChatQuery, firstProfileID, secondProfileID, "", secondProfileID).Scan(&chatID)
	if err != nil {
		return 0, err
	}

	notification := model.Notification{
		Type: "new_message",
		Payload: map[string]interface{}{
			"chat_id": 123,
			"from":    456,
			"text":    "Привет!",
		},
	}
	data, _ := json.Marshal(notification)
	cr.Client.Publish(context.Background(), "user:42:notifications", data)

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
		status,
		created_at
	FROM messages
	WHERE chat_id = $1 AND status = 2
	ORDER BY created_at ASC;
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
		if err := rows.Scan(&message.MessageID, &message.SenderID, &message.Text, &message.Status, &message.CreatedAt); err != nil {
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

	GetDeletedMessageQuery = `
		SELECT message_id, user_id, content, created_at
		FROM messages
		WHERE chat_id = $1 AND message_id = $2;
	`
)

func (cr *ChatRepo) DeleteMessage(messageID int, chatID int) error {
	tx, err := cr.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var deletedMsg model.Message
	err = tx.QueryRowContext(cr.Ctx, GetDeletedMessageQuery, chatID, messageID).Scan(
		&deletedMsg.MessageID,
		&deletedMsg.SenderID,
		&deletedMsg.Text,
		&deletedMsg.CreatedAt,
	)
	if err != nil {
		return err
	}

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

	firstID, secondID, err := cr.GetChatParticipants(chatID)
	if err != nil {
		return err
	}
	var receiverID int
	for _, uid := range []int{firstID, secondID} {
		existingMessages, err := cr.GetMessagesFromCache(chatID, uid)
		if err != nil {
			return err
		}

		var updated []model.Message
		for _, m := range existingMessages {
			if m.MessageID != messageID {
				updated = append(updated, m)
			}
		}

		if uid != deletedMsg.SenderID {
			receiverID = uid
			deletedMsg.Status = -1
			updated = append(updated, deletedMsg)
		}

		if err := cr.updateMessageCache(chatID, uid, updated); err != nil {
			return err
		}

	}

	channel := fmt.Sprintf("user:%d chat:%d messages", receiverID, chatID)
	err = cr.Client.Publish(context.Background(), channel, "new").Err()
	if err != nil {
		return err
	}

	return nil
}

const (
	InsertMessageQuery = `
		INSERT INTO messages (chat_id, user_id, content, status)
		VALUES ($1, $2, $3, $4)
		RETURNING message_id;`

	UpdateChatLastMessageQuery = `
		UPDATE chats
	SET last_message = $1, last_sender = $2
	WHERE chat_id = $3;

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

	_, err = tx.ExecContext(context.Background(), UpdateChatLastMessageQuery, content, userID, chatID)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	firstID, secondID, err := cr.GetChatParticipants(chatID)
	if err != nil {
		return 0, err
	}

	var receiverID int
	if userID == firstID {
		receiverID = secondID
	} else {
		receiverID = firstID
	}

	existingMessages, err := cr.GetMessagesFromCache(chatID, receiverID)
	if err != nil {
		return 0, err
	}

	message := model.Message{
		MessageID: messageID,
		SenderID:  userID,
		Text:      content,
		Status:    status,
		CreatedAt: time.Now(),
	}
	existingMessages = append(existingMessages, message)

	if err := cr.updateMessageCache(chatID, receiverID, existingMessages); err != nil {
		return 0, err
	}

	channel := fmt.Sprintf("user:%d chat:%d messages", receiverID, chatID)
	err = cr.Client.Publish(context.Background(), channel, "new").Err()
	if err != nil {
		return 0, err
	}

	return messageID, nil
}

const (
	UpdateMessageStatusQuery = `
		UPDATE messages
		SET status = 2
		WHERE chat_id = $1 AND user_id = $2;
	`
)

func (cr *ChatRepo) UpdateMessageStatus(chatID int, userID int) error {
	oldMessages, err := cr.GetMessagesFromCache(chatID, userID)
	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf("chat:%d:messages_user%d", chatID, userID)
	_, err = cr.Client.Del(cr.Ctx, redisKey).Result()
	if err != nil {
		return err
	}

	_, err = cr.DB.ExecContext(context.Background(), UpdateMessageStatusQuery, chatID, userID)
	if err != nil {
		return err
	}

	firstID, secondID, err := cr.GetChatParticipants(chatID)
	if err != nil {
		return err
	}

	var receiverID int
	if userID == firstID {
		receiverID = secondID
	} else {
		receiverID = firstID
	}

	existingMessages, err := cr.GetMessagesFromCache(chatID, receiverID)
	if err != nil {
		return err
	}

	for _, msg := range oldMessages {
		msg.Status = 2
		existingMessages = append(existingMessages, msg)
	}

	if err := cr.updateMessageCache(chatID, receiverID, existingMessages); err != nil {
		return err
	}

	return nil
}

func (cr *ChatRepo) GetMessagesFromCache(chatID int, userID int) ([]model.Message, error) {
	redisKey := fmt.Sprintf("chat:%d:messages_user%d", chatID, userID)

	result, err := cr.Client.Get(cr.Ctx, redisKey).Result()
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

	var filtered []model.Message
	for _, m := range messages {
		if m.Status == 1 || m.Status == -1 {
			filtered = append(filtered, m)
		}
	}

	data, _ := json.Marshal(filtered)
	cr.Client.Set(cr.Ctx, redisKey, data, 0)

	return messages, nil
}

func (cr *ChatRepo) updateMessageCache(chatID, userID int, messages []model.Message) error {
	redisKey := fmt.Sprintf("chat:%d:messages_user%d", chatID, userID)

	messageJSON, err := json.Marshal(messages)
	if err != nil {
		return err
	}
	_, err = cr.Client.Set(cr.Ctx, redisKey, messageJSON, 0).Result()
	if err != nil {
		return err
	}

	cr.Client.LTrim(cr.Ctx, redisKey, 0, 49)

	return nil
}
