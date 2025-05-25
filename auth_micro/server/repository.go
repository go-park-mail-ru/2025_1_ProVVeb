package auth

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/config"
	"github.com/go-redis/redis/v8"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type SessionRepo struct {
	DB     *sql.DB
	Client *redis.Client
	Ctx    context.Context
}

func NewSessionRepo() (*SessionRepo, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return &SessionRepo{}, err
	}

	cfg := InitPostgresConfig()
	db, err := InitPostgresConnection(cfg)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return &SessionRepo{}, err
	}

	return &SessionRepo{
		DB:     db,
		Client: client,
		Ctx:    ctx,
	}, nil
}

func CheckPostgresConfig(cfg config.DatabaseConfig) error {
	errors := map[string]struct {
		check func() bool
		msg   string
	}{
		"Host": {
			check: func() bool { return cfg.Host == "" },
			msg:   "host cannot be empty",
		},
		"Port": {
			check: func() bool { return cfg.Port < 1 || cfg.Port > 65535 },
			msg:   "invalid port number: must be between 1 and 65535",
		},
		"User": {
			check: func() bool { return cfg.User == "" },
			msg:   "user name cannot be empty",
		},
		// "Password": {
		// 	check: func() bool { return cfg.Password == "" },
		// 	msg:   "password cannot be empty",
		// },
		"DBName": {
			check: func() bool { return cfg.DBName == "" },
			msg:   "database name cannot be empty",
		},
	}

	for field, err := range errors {
		if err.check() {
			return fmt.Errorf("%s: %s", field, err.msg)
		}
	}

	return nil
}

func InitPostgresConfig() config.DatabaseConfig {
	return config.DatabaseConfig{
		Host:     "postgres",
		Port:     5432,
		User:     "app_user",
		Password: "your_secure_password",
		DBName:   "dev",
		SSLMode:  "disable",
	}
}

func InitPostgresConnection(cfg config.DatabaseConfig) (*sql.DB, error) {
	err := CheckPostgresConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %v", err)
	}

	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("error while connecting to a database: %v", err)
	}

	return db, nil
}

func ClosePostgresConnection(conn *sql.DB) error {
	var err error
	if conn != nil {
		err = conn.Close()
		if err != nil {
			fmt.Printf("failed while closing connection: %v\n", err)
		}
	}
	return err
}
