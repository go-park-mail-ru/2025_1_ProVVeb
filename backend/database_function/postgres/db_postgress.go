package postgres

import (
	"context"
	"fmt"
	"os"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/config"
	"github.com/jackc/pgx/v5"
)

func DBInitPostgresConfig() config.DatabaseConfig {
	user, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("failed to get the user home directory: %v", err))
	}

	return config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     user,
		Password: "",
		DBName:   user,
		SSLMode:  "disable",
	}
	// return config.DatabaseConfig{
	// 	Host:     "localhost",
	// 	Port:     5432,
	// 	User:     "postgres",
	// 	Password: "MYpassword",
	// 	DBName:   "dev",
	// 	SSLMode:  "disable",
	// }
}

func checkDBConfig(cfg config.DatabaseConfig) error {
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

func DBInitConnectionPostgres(cfg config.DatabaseConfig) (*pgx.Conn, error) {
	err := checkDBConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("something wrong with config ")
	}

	// connStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
	// 	cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)
	connStr := fmt.Sprintf("postgresql://%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("error while connecting to a database: %v", err)
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
