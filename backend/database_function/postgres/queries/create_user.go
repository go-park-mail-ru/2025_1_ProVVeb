package postgres

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/config"
	"github.com/jackc/pgx/v5"
)

const query_create_user = `
INSERT INTO users (login, email, phone, password, status, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	RETURNING user_id;
`

func DBCreateUserPostgres(p *pgx.Conn, user config.User) (int, error) {
	var userId int

	err := p.QueryRow(context.Background(), query_create_user, user.Login, user.Email, user.Phone, user.Password, user.Status).Scan(&userId)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return userId, nil
}
