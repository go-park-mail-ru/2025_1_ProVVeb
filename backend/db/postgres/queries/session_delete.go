package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const query_find_session = `
SELECT id FROM sessions WHERE user_id = $1;
`
const query_delete_session = `
DELETE FROM sessions WHERE user_id = $1;
`

func DBDeleteSessionPostgres(p *pgx.Conn, userID int) error {
	var profileID int
	err := p.QueryRow(context.Background(), query_find_session, userID).Scan(&profileID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("session for user with ID %d does not exist", userID)
		}
		return fmt.Errorf("failed to check if session exists: %w", err)
	}

	_, err = p.Exec(context.Background(), query_delete_session, userID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}
