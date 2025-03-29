package postgres

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/config"
	"github.com/jackc/pgx/v5"
)

const query_create_profile = `
	INSERT INTO profiles (firstname, lastname, is_male, birthday, height, description, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	RETURNING profile_id;
`

const query_create_user = `
	INSERT INTO users (login, email, phone, password, status, created_at, updated_at, profile_id)
	VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $6)
	RETURNING user_id;
`

func DBCreateUserWithProfilePostgres(p *pgx.Conn, profile config.Profile, user config.User) (int, int, error) {
	var profileId int

	err := p.QueryRow(context.Background(), query_create_profile,
		profile.FirstName, profile.LastName, profile.IsMale, profile.Birthday,
		profile.Height, profile.Description).Scan(&profileId)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create profile: %w", err)
	}

	var userId int
	err = p.QueryRow(context.Background(), query_create_user,
		user.Login, user.Email, user.Phone, user.Password, user.Status, profileId).Scan(&userId)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create user: %w", err)
	}

	return profileId, userId, nil
}
