package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const query_delete_profile = `
DELETE FROM static WHERE id = $1;
DELETE FROM profile_interests WHERE profile_id = $1;
DELETE FROM profile_preferences WHERE profile_id = $1;
DELETE FROM profile_ratings WHERE profile_id = $1 OR rated_profile_id = $1;
DELETE FROM likes WHERE profile_id = $1 OR liked_profile_id = $1;
DELETE FROM matches WHERE profile_id = $1 OR matched_profile_id = $1;
DELETE FROM messages WHERE sender_profile_id = $1 OR receiver_profile_id = $1;
DELETE FROM subscriptions WHERE profile_id = $1;
DELETE FROM complaints WHERE complaint_by = $1 OR complaint_on = $1;
DELETE FROM notifications WHERE user_id = (SELECT user_id FROM users WHERE profile_id = $1);

DELETE FROM profiles WHERE profile_id = $1;
`

const query_find_profile = `
	SELECT user_id FROM users WHERE profile_id = $1;
`

func DBDeleteProfilePostgres(p *pgx.Conn, profileID int) error {
	tx, err := p.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	var userID int
	err = tx.QueryRow(context.Background(), query_find_profile, profileID).Scan(&userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("profile with ID %d does not exist", profileID)
		}
		return fmt.Errorf("failed to check if profile exists: %w", err)
	}

	_, err = tx.Exec(context.Background(), query_delete_profile, profileID)
	if err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
