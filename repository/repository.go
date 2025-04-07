package repository

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5"
)

type UserRepository interface {
	GetUserByLogin(ctx context.Context, login string) (model.User, error)
	CreateSession(ctx context.Context, userID int, token string, expires time.Duration) (model.Session, error)
}

type SessionRepository interface {
	GetSession(sessionID string) (string, error)
	StoreSession(sessionID string, data string, ttl time.Duration) error
}

type PasswordHasher interface {
	Hash(password string) string
	Compare(hashedPassword, password string) bool
}

type UserRepo struct {
	db *pgx.Conn
}

type SessionRepo struct {
	client *redis.Client
	ctx    context.Context
}

type PassHasher struct{}

func NewUserRepo(db *pgx.Conn) *UserRepo {
	return &UserRepo{db: db}
}

func NewSessionRepo(address string, db int) *SessionRepo {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "",
		DB:       db,
	})

	ctx := context.Background()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		// логгировать ошибку подключения к Редис с печатью ошибки
		// log.Fatalf("Ошибка подключения к Redis: %v", err)
	}

	return &SessionRepo{
		client: client,
		ctx:    ctx,
	}
}

func NewPassHasher() *PassHasher {
	return &PassHasher{}
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

func (sr *SessionRepo) GetSession(sessionID string) (string, error) {
	data, err := sr.client.Get(sr.ctx, sessionID).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("сессия не найдена")
		}
		return "", fmt.Errorf("не удалось получить сессию из Redis: %v", err)
	}
	return data, nil
}

func (sr *SessionRepo) StoreSession(sessionID string, data string, ttl time.Duration) error {
	err := sr.client.Set(sr.ctx, sessionID, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("не удалось сохранить сессию в Redis: %v", err)
	}
	return nil
}

func (ph *PassHasher) Hash(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (ph *PassHasher) Compare(hashedPassword, inputPassword string) bool {
	return hashedPassword == ph.Hash(inputPassword)
}
