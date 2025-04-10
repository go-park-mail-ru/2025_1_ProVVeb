package repository

import (
	"context"
	"crypto/sha256"
	"fmt"
	"regexp"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5"
)

type UserRepository interface {
	GetUserByLogin(ctx context.Context, login string) (model.User, error)
	CreateSession(ctx context.Context, userID int, token string, expires time.Duration) (model.Session, error)
	CloseRepo() error
	UserExists(ctx context.Context, login string) bool
	StoreUser(model.User) (int, error)
	StoreProfile(model.Profile) (int, error)
}

type SessionRepository interface {
	GetSession(sessionID string) (string, error)
	StoreSession(sessionID string, data string, ttl time.Duration) error
	CloseRepo() error
}

type PasswordHasher interface {
	Hash(password string) string
	Compare(hashedPassword, password string) bool
}

type UserParamsValidator interface {
	ValidateLogin(login string) error
	ValidatePassword(password string) error
}

type UserRepo struct {
	db *pgx.Conn
}

type SessionRepo struct {
	client *redis.Client
	ctx    context.Context
}

type PassHasher struct{}

type UParamsValidator struct{}

func NewUserRepo() (*UserRepo, error) {
	cfg := InitPostgresConfig()
	db, err := InitPostgresConnection(cfg)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return &UserRepo{}, err
	}
	return &UserRepo{db: db}, nil
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

func checkPostgresConfig(cfg DatabaseConfig) error {
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

func InitPostgresConnection(cfg DatabaseConfig) (*pgx.Conn, error) {
	err := checkPostgresConfig(cfg)
	if err != nil {
		// обработать ошибку
		return nil, model.ErrInvalidUserRepoConfig
	}

	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("error while connecting to a database: %v", err)
	}

	err = conn.Ping(context.Background())
	if err != nil {
		conn.Close(context.Background())
		return nil, fmt.Errorf("failed to ping the database: %v", err)
	}

	return conn, nil
}

func ClosePostgresConnection(conn *pgx.Conn) error {
	var err error
	if conn != nil {
		err = conn.Close(context.Background())
		if err != nil {
			fmt.Printf("failed while closing connection: %v\n", err)
		}
	}
	return err
}

func NewSessionRepo(address string, db int) (*SessionRepo, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "",
		DB:       db,
	})

	ctx := context.Background()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return &SessionRepo{}, err
		// логгировать ошибку подключения к Редис с печатью ошибки
	}

	return &SessionRepo{
		client: client,
		ctx:    ctx,
	}, nil
}

func NewPassHasher() (*PassHasher, error) {
	return &PassHasher{}, nil
}

func NewUParamsValidator() (*UParamsValidator, error) {
	return &UParamsValidator{}, nil
}

const (
	getUserByLoginQuery = `
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
	createSessionQuery = `
INSERT INTO sessions (user_id, token, expires_at)
	VALUES ($1, $2, $3)
	RETURNING token, user_id, 
		EXTRACT(EPOCH FROM (expires_at - NOW()))::int
`
	createProfileQuery = `
	INSERT INTO profiles (firstname, lastname, is_male, birthday, height, description, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	RETURNING profile_id;
`
	createUserQuery = `
	INSERT INTO users (login, email, phone, password, status, created_at, updated_at, profile_id)
	VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $6)
	RETURNING user_id;
`
)

func (ur *UserRepo) GetUserByLogin(ctx context.Context, login string) (model.User, error) {
	var user model.User

	err := ur.db.QueryRow(ctx, getUserByLoginQuery, login).Scan(
		&user.UserId,
		&user.Login,
		&user.Email,
		&user.Password,
		&user.Phone,
		&user.Status,
	)

	return user, err
}

func (ur *UserRepo) CreateSession(ctx context.Context, userID int, token string, expires time.Duration) (model.Session, error) {
	var session model.Session

	err := ur.db.QueryRow(
		ctx,
		createSessionQuery,
		userID,
		token,
		time.Now().Add(expires),
	).Scan(
		&session.SessionId,
		&session.UserId,
		&session.Expires,
	)

	return session, err
}

func (ur *UserRepo) CloseRepo() error {
	return ClosePostgresConnection(ur.db)
}

func (ur *UserRepo) StoreUser(user model.User) (userId int, err error) {
	err = ur.db.QueryRow(
		context.Background(),
		createUserQuery,
		user.Login,
		user.Email,
		user.Phone,
		user.Password,
		user.Status,
		user.UserId,
	).Scan(&userId)
	return
}

func (ur *UserRepo) StoreProfile(profile model.Profile) (profileId int, err error) {
	err = ur.db.QueryRow(
		context.Background(),
		createProfileQuery,
		profile.FirstName,
		profile.LastName,
		profile.IsMale,
		profile.Birthday,
		profile.Height,
		profile.Description,
	).Scan(&profileId)
	return
}

func (ur *UserRepo) UserExists(ctx context.Context, login string) bool {
	_, err := ur.GetUserByLogin(ctx, login)
	return err == nil
}

func (sr *SessionRepo) GetSession(sessionID string) (string, error) {
	data, err := sr.client.Get(sr.ctx, sessionID).Result()
	if err != nil {
		if err == redis.Nil {
			return "", model.ErrSessionNotFound
		}
		return "", model.ErrGetSession
	}
	return data, nil
}

func (sr *SessionRepo) StoreSession(sessionID string, data string, ttl time.Duration) error {
	err := sr.client.Set(sr.ctx, sessionID, data, ttl).Err()
	if err != nil {
		return model.ErrStoreSession //
	}
	return nil
}

func (sr *SessionRepo) CloseRepo() error {
	return sr.client.Close()
}

func (ph *PassHasher) Hash(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (ph *PassHasher) Compare(hashedPassword, inputPassword string) bool {
	return hashedPassword == ph.Hash(inputPassword)
}

func (vr *UParamsValidator) ValidateLogin(login string) error {
	if (len(login) < model.MinLoginLength) || (len(login) > model.MaxLoginLength) {
		return fmt.Errorf("incorrect size of login")
	}

	re := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]*$`)
	if !re.MatchString(login) {
		return fmt.Errorf("incorrect format of login")
	}
	return nil
}

func (vr *UParamsValidator) ValidatePassword(password string) error {
	if (len(password) < model.MinPasswordLength) || (len(password) > model.MaxPasswordLength) {
		return fmt.Errorf("incorrect size of password")
	}
	// ideas for future
	// password must contain at least one digit
	// password must contain only letters and digits
	// password must contain at least one special character
	// password must not contain invalid characters

	return nil
}
