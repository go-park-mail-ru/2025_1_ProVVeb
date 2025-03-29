package postgres

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/config"
	"github.com/jackc/pgx/v5"
)

const query_get_user_by_login = `
SELECT 
    u.user_id, 
	u.login, 
    u.email,
	u.password,
    u.phone, 
    u.status
FROM users u
WHERE u.login = $1;
`

func DBGetUserPostgres(p *pgx.Conn, login string) (config.User, error) {
	var user config.User

	row := p.QueryRow(context.Background(), query_get_user_by_login, login)

	if err := row.Scan(
		&user.UserId,
		&user.Login,
		&user.Email,
		&user.Password,
		&user.Phone,
		&user.Status,
	); err != nil {
		return user, err
	}

	return user, nil
}
