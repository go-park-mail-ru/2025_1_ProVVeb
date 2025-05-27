package tests

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	mode "github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/model"
)

func setupTestDBs(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()

	connStr := "postgresql://app_user:your_secure_password@localhost:8020/dev?sslmode=disable"

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	t.Cleanup(func() { pool.Close() })

	return pool
}

func TestStoreUser_Integration(t *testing.T) {
	db := setupTestDBs(t)

	repo := &repository.UserRepo{DB: db}

	user := mode.User{UserId: 1, Login: "testuser", Email: "qwssse@sssww.ru", Password: "StrongPass123!"}
	_, err := repo.StoreUser(user)
	require.NoError(t, err)
}
