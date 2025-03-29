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

func DBCreateProfilePostgres(p *pgx.Conn, profile config.Profile) (int, error) {
	var profileId int

	err := p.QueryRow(context.Background(), query_create_profile,
		profile.FirstName, profile.LastName, profile.IsMale, profile.Birthday,
		profile.Height, profile.Description).Scan(&profileId)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return profileId, nil
}
