package repository

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type UserRepository interface {
	GetUserByLogin(login string) (model.User, error)
	StoreUser(model.User) (int, error)
	DeleteUserById(userId int) error
	UserExists(login string) bool
	StoreSession(userId int, sessionId string) error
	DeleteSession(userId int) error
	GetUserParams(userId int) (model.User, error)
	ValidateLogin(login string) error
	ValidatePassword(password string) error
	Hash(password string) string
	Compare(hashedPassword, login, password string) bool
	GetAdmin(userID int) (bool, error)

	GetPremium(userID int) (bool, int, *time.Time, error)

	CloseRepo() error
}
type DBExecutor interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type UserRepo struct {
	DB DBExecutor
}

func NewUserRepo() (*UserRepo, error) {
	cfg := InitPostgresConfig()
	db, err := InitPostgresConnection(cfg)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return nil, err
	}
	return &UserRepo{DB: db}, nil
}

func InitPostgresConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     5432,
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DB"),
		SSLMode:  os.Getenv("POSTGRES_SSLMODE"),
	}
}

func CheckPostgresConfig(cfg DatabaseConfig) error {
	errors := map[string]struct {
		check func() bool
		msg   string
	}{
		"Host": {
			check: func() bool { return cfg.Host == "" },
			msg:   "host cannot be empty",
		},
		"Port": {
			check: func() bool { return cfg.Port < 1 || cfg.Port > 65535 },
			msg:   "invalid port number: must be between 1 and 65535",
		},
		"User": {
			check: func() bool { return cfg.User == "" },
			msg:   "user name cannot be empty",
		},
		"DBName": {
			check: func() bool { return cfg.DBName == "" },
			msg:   "database name cannot be empty",
		},
	}

	for field, err := range errors {
		if err.check() {
			return fmt.Errorf("%s: %s", field, err.msg)
		}
	}

	return nil
}

func InitPostgresConnection(cfg DatabaseConfig) (*pgxpool.Pool, error) {
	err := CheckPostgresConfig(cfg)
	if err != nil {
		return nil, model.ErrInvalidUserRepoConfig
	}

	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

func ClosePostgresConnection(pool *pgxpool.Pool) error {
	if pool != nil {
		pool.Close()
	}
	return nil
}

const GetPremiumQuery = `SELECT sub_type, expires_at
FROM subscriptions
WHERE user_id = $1 AND expires_at IS NOT NULL
ORDER BY expires_at DESC
LIMIT 1;`

func (r *UserRepo) GetPremium(userID int) (bool, int, *time.Time, error) {
	var subType int
	var expiresAt *time.Time

	err := r.DB.QueryRow(context.Background(), GetPremiumQuery, userID).Scan(&subType, &expiresAt)
	if err == pgx.ErrNoRows {
		return false, 0, nil, nil
	}
	if err != nil {
		return false, 0, nil, err
	}
	if expiresAt == nil {
		return false, subType, nil, nil
	}

	return true, subType, expiresAt, nil
}

const GetUserByLoginQuery = `
SELECT 
	u.user_id, 
	u.login, 
	u.email,
	u.password,
	u.phone, 
	u.status
FROM users u
WHERE u.login = $1
AND NOT EXISTS (
    SELECT 1 FROM blacklist b WHERE b.user_id = u.user_id
  );
`

func (ur *UserRepo) GetUserByLogin(login string) (model.User, error) {
	var user model.User

	err := ur.DB.QueryRow(context.Background(), GetUserByLoginQuery, login).Scan(
		&user.UserId,
		&user.Login,
		&user.Email,
		&user.Password,
		&user.Phone,
		&user.Status,
	)

	return user, err
}

const CreateUserQuery = `
INSERT INTO users (login, email, phone, password, status, created_at, updated_at, profile_id)
VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $6)
RETURNING user_id;
`

func (ur *UserRepo) StoreUser(user model.User) (userId int, err error) {
	err = ur.DB.QueryRow(
		context.Background(),
		CreateUserQuery,
		user.Login,
		user.Email,
		user.Phone,
		user.Password,
		user.Status,
		user.UserId,
	).Scan(&userId)
	return
}

const (
	DeleteUserQuery = `
DELETE FROM users WHERE user_id = $1;
`
)

func (ur *UserRepo) DeleteUserById(userId int) error {
	_, err := ur.DB.Exec(context.Background(), DeleteUserQuery, userId)
	if err != nil {
		return model.ErrDeleteUser
	}
	return nil
}

func (ur *UserRepo) UserExists(login string) bool {
	_, err := ur.GetUserByLogin(login)
	return err == nil
}

const StoreSessionQuery = `
INSERT INTO sessions (user_id, token, created_at, expires_at)
VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + INTERVAL '72 hours')
RETURNING id;
`

func (ur *UserRepo) StoreSession(userID int, sessionID string) error {
	var sessionId int

	err := ur.DB.QueryRow(
		context.Background(),
		StoreSessionQuery,
		userID,
		sessionID,
	).Scan(&sessionId)
	return err
}

const (
	FindSessionQuery = `
SELECT id FROM sessions WHERE user_id = $1;
`
	DeleteSessionQuery = `
DELETE FROM sessions WHERE user_id = $1;
`
)

func (ur *UserRepo) DeleteSession(userId int) error {
	var profileId int
	err := ur.DB.QueryRow(context.Background(), FindSessionQuery, userId).Scan(&profileId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return model.ErrSessionNotFound
		}
		return model.ErrDeleteSession
	}

	_, err = ur.DB.Exec(context.Background(), DeleteSessionQuery, userId)
	if err != nil {
		return model.ErrDeleteSession
	}
	return err
}

const GetUserByIdQuery = `
SELECT 
	u.login, 
	u.email, 
	u.phone, 
	u.status
FROM users u
WHERE u.user_id = $1
  AND NOT EXISTS (
    SELECT 1 FROM blacklist b WHERE b.user_id = u.user_id
  );
`

func (ur *UserRepo) GetUserParams(userID int) (model.User, error) {
	var user model.User

	err := ur.DB.QueryRow(context.Background(), GetUserByIdQuery, userID).Scan(
		&user.Login,
		&user.Email,
		&user.Phone,
		&user.Status,
	)
	if err != nil {
		return user, err
	}

	return user, nil
}

const GetAdminQuery = `
	SELECT EXISTS ( SELECT 1 FROM admins WHERE user_id = $1)
`

func (ur *UserRepo) GetAdmin(userID int) (bool, error) {
	var exists bool
	err := ur.DB.QueryRow(context.Background(), GetAdminQuery, userID).Scan(&exists)
	return exists, err
}

func (ur *UserRepo) CloseRepo() error {
	pool, ok := ur.DB.(*pgxpool.Pool)
	if !ok {
		return fmt.Errorf("DBExecutor is not a *pgxpool.Pool")
	}
	return ClosePostgresConnection(pool)
}

func (ur *UserRepo) ValidateLogin(login string) error {
	if (len(login) < model.MinLoginLength) || (len(login) > model.MaxLoginLength) {
		return model.ErrInvalidLoginSize
	}

	re := regexp.MustCompile(model.ReStartsWithLetter)
	if !re.MatchString(login) {
		return model.ErrInvalidLogin
	}

	re = regexp.MustCompile(model.ReContainsLettersDigitsSymbols)
	if !re.MatchString(login) {
		return model.ErrInvalidLogin
	}

	return nil
}

func (ur *UserRepo) ValidatePassword(password string) error {
	if (len(password) < model.MinPasswordLength) || (len(password) > model.MaxPasswordLength) {
		return model.ErrInvalidPasswordSize
	}

	return nil
}

func (ur *UserRepo) Hash(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (ur *UserRepo) Compare(hashedPassword, salt, inputPassword string) bool {
	return hashedPassword == ur.Hash(salt+"_"+inputPassword)
}
