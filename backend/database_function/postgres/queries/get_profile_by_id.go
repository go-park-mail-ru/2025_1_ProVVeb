package postgres

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/config"
	"github.com/jackc/pgx/v5"
)

const query_get_profile_by_id = `
SELECT 
    p.profile_id, 
    p.firstname, 
    p.lastname, 
    p.is_male,
    p.birthday, 
    p.description, 
    l.country, 
    s.path AS avatar,
    i.description AS interest,
    pr.preference_type, 
    pr.value AS preference
FROM profiles p
LEFT JOIN locations l 
    ON p.location_id = l.location_id
LEFT JOIN static s 
    ON p.photo_id = s.id
LEFT JOIN profile_interests pi 
    ON pi.profile_id = p.profile_id
LEFT JOIN interests i 
    ON pi.interest_id = i.interest_id
LEFT JOIN profile_preferences pp 
    ON pp.profile_id = p.profile_id
LEFT JOIN preferences pr 
    ON pp.preference_id = pr.preference_id
WHERE p.profile_id = $1;
`

func DBGetProfilePostgres(p *pgx.Conn, profileID int) (config.Profile, error) {
	var profile config.Profile
	var birth time.Time
	var interest string
	var preferenceType int
	var preferenceValue string

	rows, err := p.Query(context.Background(), query_get_profile_by_id, profileID)
	if err != nil {
		return profile, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&profile.ProfileId,
			&profile.FirstName,
			&profile.LastName,
			&profile.IsMale,
			&birth,
			&profile.Description,
			&profile.Location,
			&profile.Avatar,
			&interest,
			&preferenceType,
			&preferenceValue,
		); err != nil {
			return profile, err
		}

		if interest != "" {
			profile.Interests = append(profile.Interests, interest)
		}

		if preferenceValue != "" {
			profile.Preferences = append(profile.Preferences, preferenceValue)
		}
	}

	if rows.Err() != nil {
		return profile, rows.Err()
	}

	return profile, nil
}
