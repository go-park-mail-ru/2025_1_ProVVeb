package repository

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/model"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pashagolub/pgxmock"
)

type TestUserRepoAdapter struct {
	DB pgxmock.PgxPoolIface
}

func (r *TestUserRepoAdapter) StoreUser(user model.User) (int64, error) {
	ctx := context.Background()
	_, err := r.DB.Exec(ctx,
		"INSERT INTO users \\(user_id, login, email, password\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\)",
		user.UserId, user.Login, user.Email, user.Password,
	)
	if err != nil {
		return 0, err
	}
	return int64(user.UserId), nil
}
