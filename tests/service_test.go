package tests

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/stretchr/testify/assert"
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

func InitPostgresConnection(t *testing.T) {
	repo, _ := repository.InitPostgresConnection(repository.InitPostgresConfig())
	require.NotNil(t, repo)
}

func TestHash(t *testing.T) {
	hasher, err := repository.NewPassHasher()
	require.NoError(t, err)

	password := "securePassword"
	hashedPassword := hasher.Hash(password)

	require.NotEmpty(t, hashedPassword, "hashed password should not be empty")
	require.Len(t, hashedPassword, 64, "hashed password should be 64 characters long")
}

func TestCompare(t *testing.T) {
	hasher, err := repository.NewPassHasher()
	require.NoError(t, err)

	password := "securePassword"

	hashedPassword := hasher.Hash(password)

	salt := "randomSalt"
	incorrectPassword := "incorrectPassword"
	isMatch := hasher.Compare(hashedPassword, salt, incorrectPassword)

	require.False(t, isMatch, "passwords should not match")
}

func TestCreateJwtToken(t *testing.T) {
	secret := "mySecret"
	tk, err := repository.NewJwtToken(secret)
	require.NoError(t, err)

	session := &repository.Session{
		UserID: 1,
		ID:     "session123",
	}

	token, err := tk.CreateJwtToken(session, time.Now().Add(time.Hour).Unix())
	require.NoError(t, err)
	require.NotEmpty(t, token, "token should not be empty")
}

func TestCheckJwtToken_Success(t *testing.T) {
	secret := "mySecret"
	tk, err := repository.NewJwtToken(secret)
	require.NoError(t, err)

	session := &repository.Session{
		UserID: 1,
		ID:     "session123",
	}

	token, err := tk.CreateJwtToken(session, time.Now().Add(time.Hour).Unix())
	require.NoError(t, err)

	valid, err := tk.CheckJwtToken(session, token)
	require.NoError(t, err)
	require.True(t, valid, "token should be valid")
}

func TestCheckJwtToken_Failure(t *testing.T) {
	secret := "mySecret"
	tk, err := repository.NewJwtToken(secret)
	require.NoError(t, err)

	session := &repository.Session{
		UserID: 1,
		ID:     "session123",
	}

	token, err := tk.CreateJwtToken(session, time.Now().Add(time.Hour).Unix())
	require.NoError(t, err)

	incorrectSession := &repository.Session{
		UserID: 1,
		ID:     "wrongSessionID",
	}

	valid, err := tk.CheckJwtToken(incorrectSession, token)
	require.NoError(t, err)
	require.False(t, valid, "token should be invalid")
}

func TestExtractSessionFromToken(t *testing.T) {
	secret := "mySecret"
	tk, err := repository.NewJwtToken(secret)
	require.NoError(t, err)

	session := &repository.Session{
		UserID: 1,
		ID:     "session123",
	}
	token, err := tk.CreateJwtToken(session, time.Now().Add(time.Hour).Unix())
	require.NoError(t, err)

	extractedSession, err := tk.ExtractSessionFromToken(token)
	require.NoError(t, err)
	require.NotNil(t, extractedSession)
	require.Equal(t, session.ID, extractedSession.ID, "session IDs should match")
	require.Equal(t, session.UserID, extractedSession.UserID, "user IDs should match")
}

func TestInitPostgresConfigProf(t *testing.T) {
	cfg := repository.InitPostgresConfig()
	assert.Equal(t, "postgres", cfg.Host)
	assert.Equal(t, 5432, cfg.Port)
	assert.Equal(t, "app_user", cfg.User)
	assert.Equal(t, "your_secure_password", cfg.Password)
	assert.Equal(t, "dev", cfg.DBName)
	assert.Equal(t, "disable", cfg.SSLMode)
}

func TestCheckPostgresConfigProf(t *testing.T) {
	tests := []struct {
		name    string
		cfg     repository.DatabaseConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: repository.DatabaseConfig{
				Host:    "localhost",
				Port:    5432,
				User:    "postgres",
				DBName:  "dev",
				SSLMode: "disable",
			},
			wantErr: false,
		},
		{
			name: "missing host",
			cfg: repository.DatabaseConfig{
				Host:    "",
				Port:    5432,
				User:    "postgres",
				DBName:  "dev",
				SSLMode: "disable",
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			cfg: repository.DatabaseConfig{
				Host:    "localhost",
				Port:    99999,
				User:    "postgres",
				DBName:  "dev",
				SSLMode: "disable",
			},
			wantErr: true,
		},
		{
			name: "missing user",
			cfg: repository.DatabaseConfig{
				Host:    "localhost",
				Port:    5432,
				User:    "",
				DBName:  "dev",
				SSLMode: "disable",
			},
			wantErr: true,
		},
		{
			name: "missing db name",
			cfg: repository.DatabaseConfig{
				Host:    "localhost",
				Port:    5432,
				User:    "postgres",
				DBName:  "",
				SSLMode: "disable",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		err := repository.CheckPostgresConfig(tt.cfg)
		if tt.wantErr {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
		}
	}
}

func TestInitPostgresConnection_InvalidConfig(t *testing.T) {
	cfg := repository.DatabaseConfig{
		Host:    "",
		Port:    5432,
		User:    "postgres",
		DBName:  "dev",
		SSLMode: "disable",
	}

	db, err := repository.InitPostgresConnection(cfg)
	assert.Nil(t, db)
	assert.True(t, errors.Is(err, model.ErrInvalidUserRepoConfig))
}

func TestClosePostgresConnection(t *testing.T) {
	err := repository.ClosePostgresConnection(nil)
	assert.NoError(t, err)

	db, _ := sql.Open("pgx", "postgresql://invalid:invalid@localhost:5432/invalid?sslmode=disable")
	_ = db.Close()
	err = repository.ClosePostgresConnection(db)
	assert.Nil(t, err)
}
