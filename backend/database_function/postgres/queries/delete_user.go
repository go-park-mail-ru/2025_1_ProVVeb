package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const query_delete_user = `
DELETE FROM static WHERE id = $1;
DELETE FROM profile_interests WHERE profile_id = $1;
DELETE FROM profile_preferences WHERE profile_id = $1;
DELETE FROM profile_ratings WHERE profile_id = $1 OR rated_profile_id = $1;
DELETE FROM sessions WHERE user_id = $1;
DELETE FROM likes WHERE profile_id = $1 OR liked_profile_id = $1;
DELETE FROM matches WHERE profile_id = $1 OR matched_profile_id = $1;
DELETE FROM subscriptions WHERE user_id = $1;
DELETE FROM complaints WHERE complaint_by = $1 OR complaint_on = $1;
DELETE FROM blacklist WHERE user_id = $1;
DELETE FROM notifications WHERE user_id = $1;

DELETE FROM users WHERE user_id = $1;
`

const query_find_user = `
	SELECT profile_id FROM users WHERE user_id = $1;
`

func DBDeleteUserPostgres(p *pgx.Conn, userID int) error {
	tx, err := p.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	var profileID int
	err = tx.QueryRow(context.Background(), query_find_user, userID).Scan(&profileID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("user with ID %d does not exist", userID)
		}
		return fmt.Errorf("failed to check if user exists: %w", err)
	}

	_, err = tx.Exec(context.Background(), query_delete_user, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
