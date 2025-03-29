package postgres

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/config"
	"github.com/jackc/pgx/v5"
)

func DBInitPostgresConfig() config.DatabaseConfig {
	return config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "Grey31415",
		DBName:   "dev",
		SSLMode:  "disable",
	}
}

func DBInitConnectionPostgres(cfg config.DatabaseConfig) (*pgx.Conn, error) {
	if cfg.DBName == "" {
		return nil, fmt.Errorf("database name cannot be empty")
	}

	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("error while connecting to a database")
	}

	err = conn.Ping(context.Background())
	if err != nil {
		conn.Close(context.Background())
		return nil, fmt.Errorf("failed to ping the database: %v", err)
	}

	return conn, nil
}

func DBCloseConnectionPostgres(conn *pgx.Conn) {
	if conn != nil {
		err := conn.Close(context.Background())
		if err != nil {
			fmt.Printf("Ошибка при закрытии соединения: %v\n", err)
		}
	}
}
