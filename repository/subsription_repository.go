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
	CreateSub(userID int, sub_type int) error
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
INSERT INTO subscriptions (user_id, sub_type, expires_at)
VALUES (
    $1,
    2,
    NOW() + make_interval(days := 3 + 30 * $2)
);
`

func (sr *SubRepo) CreateSub(userID int, subType int) error {
	_, err := sr.DB.ExecContext(context.Background(), CreateSubQuery, userID, subType)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}
	return nil
}
