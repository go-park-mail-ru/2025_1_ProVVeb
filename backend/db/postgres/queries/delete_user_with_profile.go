package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const query_find_user_profile = `
SELECT profile_id FROM users WHERE user_id = $1;
`
const query_delete_user = `
DELETE FROM users WHERE user_id = $1;
`

const query_delete_profile = `
DELETE FROM profiles WHERE profile_id = $1;
`

func DBDeleteUserWithProfilePostgres(p *pgx.Conn, userID int) error {
	var profileID int
	err := p.QueryRow(context.Background(), query_find_user_profile, userID).Scan(&profileID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("user with ID %d does not exist", userID)
		}
		return fmt.Errorf("failed to check if user exists: %w", err)
	}

	_, err = p.Exec(context.Background(), query_delete_profile, profileID)
	if err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	_, err = p.Exec(context.Background(), query_delete_user, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
