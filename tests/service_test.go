package tests

import (
	"testing"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/stretchr/testify/require"
)

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

func TestNewSessionRepo(t *testing.T) {
	repo, _ := repository.NewSessionRepo()
	require.NotNil(t, repo)
}

func NewStaticRepo(t *testing.T) {
	repo, _ := repository.NewStaticRepo()
	require.NotNil(t, repo)
}

func NewUserRepo(t *testing.T) {
	repo, _ := repository.NewUserRepo()
	require.NotNil(t, repo)
}

func InitPostgresConnection(t *testing.T) {
	repo, _ := repository.InitPostgresConnection(repository.InitPostgresConfig())
	require.NotNil(t, repo)
}
