package postgres

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
)

const query_create_session = `
	INSERT INTO sessions (user_id, token, created_at, expires_at)
	VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + INTERVAL '72 hours')
	RETURNING id;
`

func DBCreateSessionPostgres(p *pgx.Conn, cookie http.Cookie, UserId int) (int, error) {
	var sessionId int

	err := p.QueryRow(context.Background(), query_create_session,
		UserId, cookie.Value).Scan(&sessionId)
	if err != nil {
		return 0, fmt.Errorf("failed to create session: %w", err)
	}

	return sessionId, nil
}
