package repository

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"regexp"
	"slices"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type UserRepository interface {
	GetUserByLogin(ctx context.Context, login string) (model.User, error)
	CreateSession(ctx context.Context, userID int, token string, expires time.Duration) (model.Session, error)
	CloseRepo() error
	UserExists(ctx context.Context, login string) bool
	StoreUser(model.User) (int, error)
	StoreProfile(model.Profile) (int, error)
	DeleteSession(userId int) error
	StoreSession(userID int, sessionID string) error
	DeleteUserById(userId int) error
	GetProfileById(userId int) (model.Profile, error)
	GetProfilesByUserId(forUserId int) ([]model.Profile, error)
	SetLike(from int, to int, status int) (likeID int, err error)

	StorePhoto(userID int, url string) error
	GetPhotos(userID int) ([]string, error)
	GetMatches(forUserId int) ([]model.Profile, error)
}

type SessionRepository interface {
	GetSession(sessionID string) (string, error)
	StoreSession(sessionID string, data string, ttl time.Duration) error
	CloseRepo() error
	DeleteSession(sessionID string) error
}

type PasswordHasher interface {
	Hash(password string) string
	Compare(hashedPassword, login, password string) bool
}

type UserParamsValidator interface {
	ValidateLogin(login string) error
	ValidatePassword(password string) error
}

type StaticRepository interface {
	GetImages(urls []string) ([][]byte, error)
	UploadImages(fileBytes []byte, filename, contentType string) error
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

type StaticRepo struct {
	client     *minio.Client
	bucketName string
}

const getMatches = `
SELECT 
    profile_id, 
    matched_profile_id
FROM matches
WHERE profile_id = $1 OR matched_profile_id = $1;
`

func (ur *UserRepo) GetMatches(forUserId int) ([]model.Profile, error) {
	rows, err := ur.db.Query(context.Background(), getMatches, forUserId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches [][2]int

	for rows.Next() {
		var a, b int
		if err := rows.Scan(&a, &b); err != nil {
			return nil, err
		}
		matches = append(matches, [2]int{a, b})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	profiles := make([]model.Profile, 0, model.PageSize)
	amount := 0
	for i := 0; amount < len(matches); i++ {
		targ := matches[i][0]
		if matches[i][0] == forUserId {
			targ = matches[i][1]
		}
		profile, err := ur.GetProfileById(targ)
		if err != nil {
			return profiles, err
		}
		profiles = append(profiles, profile)
		amount++
	}
	return profiles, nil
}

func (sr *StaticRepo) UploadImages(fileBytes []byte, filename, contentType string) error {
	ctx := context.Background()

	_, err := sr.client.PutObject(ctx, sr.bucketName, filename,
		bytes.NewReader(fileBytes),
		int64(len(fileBytes)),
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return fmt.Errorf("failed to upload image to minio: %w", err)
	}
	return nil
}

func (sr *StaticRepo) GetImages(urls []string) ([][]byte, error) {
	var results [][]byte

	for _, objectName := range urls {
		obj, err := sr.client.GetObject(context.Background(), sr.bucketName, objectName, minio.GetObjectOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get object %s: %w", objectName, err)
		}

		data, err := io.ReadAll(obj)
		if err != nil {
			return nil, fmt.Errorf("failed to read object %s: %w", objectName, err)
		}

		results = append(results, data)
	}

	return results, nil
}

// endpoint := "213.219.214.83:8030"

func NewStaticRepo() (*StaticRepo, error) {
	endpoint := "minio:9000"
	accessKeyID := "minioadmin"
	secretAccessKey := "miniopassword"
	useSSL := false

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return &StaticRepo{}, err
	}

	bucketName := "profile-photos"
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, err
	}
	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
	}

	return &StaticRepo{
		client:     minioClient,
		bucketName: bucketName,
	}, nil
}

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

func NewSessionRepo() (*SessionRepo, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
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

const getUserByLoginQuery = `
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

const createSessionQuery = `
INSERT INTO sessions (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING token, user_id, 
	EXTRACT(EPOCH FROM (expires_at - NOW()))::int
`

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

const storeSessionQuery = `
INSERT INTO sessions (user_id, token, created_at, expires_at)
VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + INTERVAL '72 hours')
RETURNING id;
`

func (ur *UserRepo) StoreSession(userID int, sessionID string) error {
	var sessionId int

	err := ur.db.QueryRow(
		context.Background(),
		storeSessionQuery,
		userID,
		sessionID,
	).Scan(&sessionId)
	return err
}

func (ur *UserRepo) CloseRepo() error {
	return ClosePostgresConnection(ur.db)
}

const createUserQuery = `
INSERT INTO users (login, email, phone, password, status, created_at, updated_at, profile_id)
VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $6)
RETURNING user_id;
`

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

const createProfileQuery = `
INSERT INTO profiles (firstname, lastname, is_male, birthday, height, description, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING profile_id;
`

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

const (
	findSessionQuery = `
SELECT id FROM sessions WHERE user_id = $1;
`
	deleteSessionQuery = `
DELETE FROM sessions WHERE user_id = $1;
`
)

func (ur *UserRepo) DeleteSession(userId int) error {
	var profileId int
	err := ur.db.QueryRow(context.Background(), findSessionQuery, userId).Scan(&profileId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return model.ErrSessionNotFound
		}
		return model.ErrDeleteSession
	}

	_, err = ur.db.Exec(context.Background(), deleteSessionQuery, userId)
	if err != nil {
		return model.ErrDeleteSession
	}
	return err
}

const (
	deleteProfileQuery = `
DELETE FROM profiles WHERE profile_id = $1;
`
	deleteUserQuery = `
DELETE FROM users WHERE user_id = $1;
`
	findUserProfileQuery = `
	SELECT profile_id FROM users WHERE user_id = $1;
	`
)

func (ur *UserRepo) DeleteUserById(userId int) error {
	var profileId int
	err := ur.db.QueryRow(context.Background(), findUserProfileQuery, userId).Scan(&profileId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return model.ErrProfileNotFound
		}
		return model.ErrDeleteUser
	}

	_, err = ur.db.Exec(context.Background(), deleteProfileQuery, profileId)
	if err != nil {
		return model.ErrDeleteProfile
	}

	_, err = ur.db.Exec(context.Background(), deleteUserQuery, userId)
	if err != nil {
		return model.ErrDeleteUser
	}
	return nil
}

const getProfileByIdQuery = `
SELECT 
    p.profile_id, 
    p.firstname, 
    p.lastname, 
    p.is_male,
    p.height,
    p.birthday, 
    p.description, 
    l.country, 
    liked.profile_id AS liked_by_profile_id,
    s.path AS avatar,
    i.description AS interest,
    pr.preference_description || ':' || pr.preference_value AS preference
FROM profiles p
LEFT JOIN locations l 
    ON p.location_id = l.location_id
LEFT JOIN "static" s 
    ON p.profile_id = s.profile_id
LEFT JOIN profile_interests pi 
    ON pi.profile_id = p.profile_id
LEFT JOIN interests i 
    ON pi.interest_id = i.interest_id
LEFT JOIN profile_preferences pp 
    ON pp.profile_id = p.profile_id
LEFT JOIN preferences pr 
    ON pp.preference_id = pr.preference_id
LEFT JOIN likes liked
    ON liked.liked_profile_id = p.profile_id
WHERE p.profile_id = $1;

`

func (ur *UserRepo) GetProfileById(profileId int) (model.Profile, error) {
	var profile model.Profile
	var birth sql.NullTime
	var interest sql.NullString
	var preferenceValue sql.NullString
	var likedByProfileId sql.NullInt64
	var photo sql.NullString
	var location sql.NullString

	rows, err := ur.db.Query(context.Background(), getProfileByIdQuery, profileId)
	if err != nil {
		return profile, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&profile.ProfileId,
			&profile.FirstName,
			&profile.LastName,
			&profile.IsMale,
			&profile.Height,
			&birth,
			&profile.Description,
			&location,
			&likedByProfileId,
			&photo,
			&interest,
			&preferenceValue,
		); err != nil {
			return profile, err
		}

		if photo.Valid {
			profile.Card = "http://213.219.214.83:8080/static/cards" + photo.String
			profile.Avatar = "http://213.219.214.83:8080/static/avatars" + photo.String
		} else {
			profile.Avatar = ""
			profile.Card = ""
		}

		if birth.Valid {
			profile.Birthday = birth.Time
		}

		if location.Valid {
			profile.Location = location.String
		}
		if likedByProfileId.Valid && !slices.Contains(profile.LikedBy, int(likedByProfileId.Int64)) {
			profile.LikedBy = append(profile.LikedBy, int(likedByProfileId.Int64))
		}

		if interest.Valid && !slices.Contains(profile.Interests, interest.String) {
			profile.Interests = append(profile.Interests, interest.String)
		}

		if preferenceValue.Valid && !slices.Contains(profile.Preferences, preferenceValue.String) {
			profile.Preferences = append(profile.Preferences, preferenceValue.String)
		}
	}
	if rows.Err() != nil {
		return profile, rows.Err()
	}

	return profile, nil
}

const createLikeQuery = `
INSERT INTO likes (profile_id, liked_profile_id, created_at, status)
VALUES ($1, $2, CURRENT_TIMESTAMP, $3)
RETURNING like_id;
`

func (ur *UserRepo) SetLike(from int, to int, status int) (likeID int, err error) {
	err = ur.db.QueryRow(
		context.Background(),
		createLikeQuery,
		from,
		to,
		status,
	).Scan(&likeID)
	return
}

const uploadPhotoQuery = `
INSERT INTO static (profile_id, path, created_at, updated_at)
VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING profile_id, path, created_at;
`

func (ur *UserRepo) StorePhoto(userID int, url string) error {
	_, err := ur.db.Exec(context.Background(), uploadPhotoQuery, userID, url)
	return err
}

const getPhotoPathsQuery = `
SELECT path FROM static 
WHERE profile_id = (
	SELECT profile_id FROM users WHERE user_id = $1
);
`

func (ur *UserRepo) GetPhotos(userID int) ([]string, error) {
	rows, err := ur.db.Query(context.Background(), getPhotoPathsQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		photos = append(photos, path)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return photos, nil
}

func (ur *UserRepo) GetProfilesByUserId(forUserId int) ([]model.Profile, error) {
	profiles := make([]model.Profile, 0, model.PageSize)
	amount := 0
	for i := 0; amount < model.PageSize; i++ {
		if i != forUserId {
			profile, err := ur.GetProfileById(i)
			if err != nil {
				return profiles, err
			}
			profiles = append(profiles, profile)
			amount++
		}
	}
	return profiles, nil
}

func (sr *SessionRepo) DeleteSession(sessionID string) error {
	return sr.client.Del(sr.ctx, sessionID).Err()
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
		return model.ErrStoreSession
	}
	return nil
}

func (sr *SessionRepo) DeleteAllSessions() error {
	return sr.client.FlushAll(sr.ctx).Err()
}

func (sr *SessionRepo) CloseRepo() error {
	return sr.client.Close()
}

func (ph *PassHasher) Hash(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (ph *PassHasher) Compare(hashedPassword, inputLogin, inputPassword string) bool {
	return hashedPassword == ph.Hash(inputLogin+"_"+inputPassword)
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
