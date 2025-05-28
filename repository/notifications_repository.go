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

type NotificationsRepository interface {
	GetNotifications(userID int) ([]model.NotificationSend, error)
	MarkNotifications(userID int, nofit_type string) error
	DeleteNotifications(notification_id int, userID int) error

	GetCurrentNotifications(userID int) ([]model.NotificationSend, error)
	AddNotification(userID int, notif model.NotificationSend) error
}

type RedisClient interface {
	LRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd
	TxPipeline() redis.Pipeliner
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Close() error
	Ping(ctx context.Context) *redis.StatusCmd
	Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd
}

type NotificationsRepo struct {
	DB     *sql.DB
	Client RedisClient
	Ctx    context.Context
}

func NewNotificationsRepo() (*NotificationsRepo, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return &NotificationsRepo{}, err
	}

	cfg := InitPostgresConfig()
	db, err := InitPostgresConnection(cfg)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return &NotificationsRepo{}, err
	}

	return &NotificationsRepo{
		DB:     db,
		Client: client,
		Ctx:    ctx,
	}, nil
}

func (nr *NotificationsRepo) CloseRepo() error {
	return nr.Client.Close()
}

const GetNotificationQuery = `
SELECT
    n.notification_id,
    nt.type_description,
    n.content,
	n.created_at,
	n.read_at
FROM notifications n
JOIN notification_types nt ON n.notification_type = nt.notif_type
WHERE n.user_id = $1
ORDER BY n.created_at DESC;
`

func (nr *NotificationsRepo) GetNotifications(userID int) ([]model.NotificationSend, error) {
	rows, err := nr.DB.QueryContext(context.Background(), GetNotificationQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []model.NotificationSend
	var CreatedAt time.Time
	for rows.Next() {
		var notif model.NotificationSend
		var ReadAt sql.NullTime
		if err := rows.Scan(&notif.NotificationID, &notif.NotifType, &notif.Content, &CreatedAt, &ReadAt); err != nil {
			return nil, err
		}
		if !ReadAt.Valid {
			notif.Read = 0
		} else {
			notif.Read = 1
		}
		notifications = append(notifications, notif)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

const UpdateNotifications = `
UPDATE notifications n
SET read_at = CURRENT_TIMESTAMP
FROM notification_types nt
WHERE n.notification_type = nt.notif_type
  AND n.user_id = $1
  AND nt.type_description = $2;
`

func (nr *NotificationsRepo) MarkNotifications(userID int, notifType string) error {
	_, err := nr.DB.ExecContext(context.Background(), UpdateNotifications, userID, notifType)
	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf("CACHE:user:%dnotifications", userID)
	fmt.Println(redisKey)

	items, err := nr.Client.LRange(nr.Ctx, redisKey, 0, 99).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return err
	}

	notifications := make([]string, 0, len(items))
	for _, item := range items {
		var notif model.NotificationSend
		if err := json.Unmarshal([]byte(item), &notif); err != nil {
			continue
		}
		if notif.NotifType == notifType {
			notif.Read = 1
		}
		updated, err := json.Marshal(notif)
		if err != nil {
			continue
		}
		notifications = append(notifications, string(updated))
	}

	pipe := nr.Client.TxPipeline()
	pipe.Del(nr.Ctx, redisKey)
	if len(notifications) > 0 {
		pipe.RPush(nr.Ctx, redisKey, notifications)
	}
	_, err = pipe.Exec(nr.Ctx)
	return err
}

const DeleteNotification = `
DELETE FROM notifications
WHERE notification_id = $1 AND user_id = $2;
`

func (nr *NotificationsRepo) DeleteNotifications(notification_id int, userID int) error {
	_, err := nr.DB.ExecContext(context.Background(), DeleteNotification, notification_id, userID)
	if err != nil {
		return err
	}
	redisKey := fmt.Sprintf("CACHE:user:%dnotifications", userID)
	_, err = nr.Client.Del(nr.Ctx, redisKey).Result()
	if err != nil {
		return err
	}

	return nil
}

func (nr *NotificationsRepo) GetCurrentNotifications(userID int) ([]model.NotificationSend, error) {
	redisKey := fmt.Sprintf("CACHE:user:%dnotifications", userID)
	fmt.Println(redisKey)

	items, err := nr.Client.LRange(nr.Ctx, redisKey, 0, 99).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	notifications := make([]model.NotificationSend, 0, len(items))
	for _, item := range items {
		var notif model.NotificationSend
		if err := json.Unmarshal([]byte(item), &notif); err != nil {
			continue
		}
		notifications = append(notifications, notif)
	}

	return notifications, nil
}

const AddNotificationQuery = `
INSERT INTO notifications (user_id, notification_type, content)
VALUES (
    $1,
    (SELECT notif_type FROM notification_types WHERE type_description = $2),
    $3
)
RETURNING notification_id;
`

func (nr *NotificationsRepo) AddNotification(userID int, notif model.NotificationSend) error {

	var notifID int
	err := nr.DB.QueryRowContext(
		context.Background(),
		AddNotificationQuery,
		userID,
		notif.NotifType,
		notif.Content,
	).Scan(&notifID)
	if err != nil {
		return err
	}

	notif.NotificationID = notifID

	redisKey := fmt.Sprintf("CACHE:user:%dnotifications", userID)
	jsonNotif, err := json.Marshal(notif)
	if err != nil {
		return err
	}

	pipe := nr.Client.TxPipeline()
	pipe.LPush(nr.Ctx, redisKey, jsonNotif)
	pipe.LTrim(nr.Ctx, redisKey, 0, 99)
	pipe.Expire(nr.Ctx, redisKey, 30*time.Hour)
	if _, err := pipe.Exec(nr.Ctx); err != nil {
		return err
	}

	channel := fmt.Sprintf("user:%d notifications", userID)
	if err := nr.Client.Publish(context.Background(), channel, "new").Err(); err != nil {
		return err
	}

	return nil
}
