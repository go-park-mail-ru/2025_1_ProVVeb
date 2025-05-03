package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"regexp"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/model"
	"github.com/jackc/pgx/v5"
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

	CloseRepo() error
}

type UserRepo struct {
	DB *sql.DB
}

func NewUserRepo() (*UserRepo, error) {
	cfg := InitPostgresConfig()
	db, err := InitPostgresConnection(cfg)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return &UserRepo{}, err
	}
	return &UserRepo{DB: db}, nil
}

func InitPostgresConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     "postgres",
		Port:     5432,
		User:     "postgres",
		Password: "Grey31415",
		DBName:   "dev",
		SSLMode:  "disable",
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
		// "Password": {
		// 	check: func() bool { return cfg.Password == "" },
		// 	msg:   "password cannot be empty",
		// },
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

func InitPostgresConnection(cfg DatabaseConfig) (*sql.DB, error) {
	err := CheckPostgresConfig(cfg)
	if err != nil {
		return nil, model.ErrInvalidUserRepoConfig
	}

	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("error while connecting to a database: %v", err)
	}

	return db, nil
}

func ClosePostgresConnection(conn *sql.DB) error {
	var err error
	if conn != nil {
		err = conn.Close()
		if err != nil {
			fmt.Printf("failed while closing connection: %v\n", err)
		}
	}
	return err
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
WHERE u.login = $1;
`

func (ur *UserRepo) GetUserByLogin(login string) (model.User, error) {
	var user model.User

	err := ur.DB.QueryRowContext(context.Background(), GetUserByLoginQuery, login).Scan(
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
	err = ur.DB.QueryRowContext(
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
	_, err := ur.DB.ExecContext(context.Background(), DeleteUserQuery, userId)
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

	err := ur.DB.QueryRowContext(
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
	err := ur.DB.QueryRowContext(context.Background(), FindSessionQuery, userId).Scan(&profileId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return model.ErrSessionNotFound
		}
		return model.ErrDeleteSession
	}

	_, err = ur.DB.ExecContext(context.Background(), DeleteSessionQuery, userId)
	if err != nil {
		return model.ErrDeleteSession
	}
	return err
}

const GetUserByIdQuery = `
	SELECT login, email, phone, status FROM users WHERE user_id = $1;
`

func (ur *UserRepo) GetUserParams(userID int) (model.User, error) {
	var user model.User

	err := ur.DB.QueryRowContext(context.Background(), GetUserByIdQuery, userID).Scan(
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
	err := ur.DB.QueryRowContext(context.Background(), GetAdminQuery, userID).Scan(&exists)
	return exists, err
}

func (ur *UserRepo) CloseRepo() error {
	return ClosePostgresConnection(ur.DB)
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
