package tests

import (
	"testing"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/stretchr/testify/require"
)

func TestNewUserRepo(t *testing.T) {
	repo, err := repository.NewUserRepo()
	require.NotNil(t, repo)
	require.Nil(t, err)
}

func TestCheckPostgresConfig(t *testing.T) {
	cfg := repository.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "test",
		Password: "test",
		DBName:   "testdb",
	}

	err := repository.CheckPostgresConfig(cfg)
	require.NoError(t, err)

	cfg.Host = ""
	err = repository.CheckPostgresConfig(cfg)
	require.Error(t, err)
}

func TestInitPostgresConfig(t *testing.T) {
	connStr := repository.InitPostgresConfig()
	require.NotNil(t, connStr)
}
