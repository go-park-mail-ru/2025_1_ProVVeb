package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type SubsriptionRepository interface {
	CreateSub(userID int, subType int, data string) error
	UpdateBorder(userID int, new_border int) error
}

type SubRepo struct {
	DB  *sql.DB
	Ctx context.Context
}

func NewSubRepo() (*SubRepo, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return &SubRepo{}, err
	}

	cfg := InitPostgresConfig()
	db, err := InitPostgresConnection(cfg)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return &SubRepo{}, err
	}

	return &SubRepo{
		DB:  db,
		Ctx: ctx,
	}, nil
}

const CreateSubQuery = `
INSERT INTO subscriptions (user_id, sub_type, expires_at, transaction_data)
VALUES (
    $1,
    2,
    NOW() + make_interval(days := 3 + 30 * $2),
    $3
)
ON CONFLICT (user_id) DO UPDATE
SET 
    sub_type = EXCLUDED.sub_type,
    expires_at = EXCLUDED.expires_at,
    transaction_data = EXCLUDED.transaction_data,
    created_at = CURRENT_TIMESTAMP;

`

func (sr *SubRepo) CreateSub(userID int, subType int, data string) error {
	_, err := sr.DB.ExecContext(context.Background(), CreateSubQuery, userID, subType, data)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}
	return nil
}

const UpdateBorderQuery = `
UPDATE subscriptions
		SET border = $1
		WHERE user_id = $2;

`

func (sr *SubRepo) UpdateBorder(userID int, new_border int) error {
	_, err := sr.DB.ExecContext(context.Background(), UpdateBorderQuery, new_border, userID)
	if err != nil {
		return fmt.Errorf("failed to update border: %w", err)
	}
	return nil
}
