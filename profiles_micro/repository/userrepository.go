package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/model"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type ProfileRepository interface {
	GetProfileById(userId int) (model.Profile, error)
	StoreProfile(model.Profile) (int, error)
	GetProfilesByUserId(forUserId int) ([]model.Profile, error)
	GetMatches(forUserId int) ([]model.Profile, error)
	UpdateProfile(int, model.Profile) error
	GetPhotos(userId int) ([]string, error)
	DeletePhoto(userId int, url string) error
	StorePhoto(userId int, url string) error
	DeleteProfile(userId int) error
	StorePhotos(profileId int, paths []string) error
	StoreInterests(profileId int, interests []string) error
	SetLike(from int, to int, status int) (int, error)
	SearchProfiles(cur_user int, params model.SearchProfileRequest) ([]model.FoundProfile, error)
	GetProfileStats(profileID int) (model.ProfileStats, error)
	GetRecomendations(profileId int) (model.Profile, error)
	CloseRepo()
}

type DBQuerier interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row

	Begin(ctx context.Context) (pgx.Tx, error)
}

type ProfileRepo struct {
	DB     DBQuerier
	Client *redis.Client
}

func NewUserRepo() (*ProfileRepo, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return &ProfileRepo{}, err
	}

	cfg := InitPostgresConfig()
	db, err := InitPostgresConnection(cfg)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return &ProfileRepo{}, err
	}

	return &ProfileRepo{
		DB:     db,
		Client: client,
	}, nil
}

func InitPostgresConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     5432,
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DB"),
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

func ClosePostgresConnection(pool *pgxpool.Pool) {
	if pool != nil {
		pool.Close()
	}
}

const GetProfileByIdQuery = `
WITH base_profile AS (
    SELECT 
        profile_id, firstname, lastname, is_male,
        height, birthday, description, goal, location_id
    FROM profiles
    WHERE profile_id = $1
)
SELECT 
    bp.profile_id,
    bp.firstname,
    bp.lastname,
    bp.is_male,
    bp.height,
    bp.birthday,
    bp.description,
    bp.goal,
    l.country,
    l.city,
    l.district,
    liked.profile_id AS liked_by_profile_id,
    s.path AS avatar,
    i.description AS interest,
    pr.preference_description,
    pr.preference_value,
    param.parameter_description,
    param.parameter_value,
	CASE WHEN sbs.sub_id IS NOT NULL THEN TRUE ELSE FALSE END AS premium_status,
    sbs.border
FROM base_profile bp
LEFT JOIN locations l ON bp.location_id = l.location_id
LEFT JOIN "static" s ON bp.profile_id = s.profile_id
LEFT JOIN profile_interests pi ON pi.profile_id = bp.profile_id
LEFT JOIN interests i ON pi.interest_id = i.interest_id
LEFT JOIN profile_preferences pp ON pp.profile_id = bp.profile_id
LEFT JOIN preferences pr ON pp.preference_id = pr.preference_id
LEFT JOIN profile_parameter pp2 ON pp2.profile_id = bp.profile_id
LEFT JOIN parameters param ON pp2.parameter_id = param.parameter_id
LEFT JOIN likes liked ON liked.liked_profile_id = bp.profile_id
LEFT JOIN subscriptions sbs ON sbs.user_id = bp.profile_id AND sbs.expires_at > NOW();

`

func (pr *ProfileRepo) GetProfileById(profileId int) (model.Profile, error) {
	var profile model.Profile
	var birth sql.NullTime
	var interest sql.NullString
	var preferenceDesc sql.NullString
	var preferenceValue sql.NullString
	var likedByProfileId sql.NullInt64
	var photo sql.NullString
	var country, city, district sql.NullString

	var goal sql.NullInt64
	var paramDesc, paramValue sql.NullString

	var premiumStatus sql.NullBool
	var premiumBorder sql.NullInt64

	ctx := context.Background()
	rows, err := pr.DB.Query(ctx, GetProfileByIdQuery, profileId)

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
			&goal,
			&country,
			&city,
			&district,
			&likedByProfileId,
			&photo,
			&interest,
			&preferenceDesc,
			&preferenceValue,
			&paramDesc,
			&paramValue,
			&premiumStatus,
			&premiumBorder,
		); err != nil {
			return profile, err
		}

		if birth.Valid {
			profile.Birthday = birth.Time
		}

		if country.Valid && city.Valid && district.Valid {
			profile.Location = fmt.Sprintf("%s@%s@%s", country.String, city.String, district.String)
		}

		if likedByProfileId.Valid && !slices.Contains(profile.LikedBy, int(likedByProfileId.Int64)) {
			profile.LikedBy = append(profile.LikedBy, int(likedByProfileId.Int64))
		}

		if interest.Valid && !slices.Contains(profile.Interests, interest.String) {
			profile.Interests = append(profile.Interests, interest.String)
		}
		if goal.Valid {
			profile.Goal = int(goal.Int64)
		}

		if paramDesc.Valid && paramValue.Valid {
			param := model.Preference{
				Description: paramDesc.String,
				Value:       paramValue.String,
			}
			if !slices.Contains(profile.Parameters, param) {
				profile.Parameters = append(profile.Parameters, param)
			}
		}

		if preferenceDesc.Valid && preferenceValue.Valid {
			pref := model.Preference{
				Description: preferenceDesc.String,
				Value:       preferenceValue.String,
			}
			if !slices.Contains(profile.Preferences, pref) {
				profile.Preferences = append(profile.Preferences, pref)
			}
		}
		if photo.Valid && !slices.Contains(profile.Photos, photo.String) {
			profile.Photos = append(profile.Photos, photo.String)
		}
		if premiumStatus.Valid && premiumStatus.Bool {
			profile.Premium.Status = true
			if premiumBorder.Valid {
				profile.Premium.Border = int(premiumBorder.Int64)
			}
		}
	}

	if rows.Err() != nil {
		return profile, rows.Err()
	}

	return profile, nil
}

const CreateProfileQuery = `
INSERT INTO profiles (
    firstname, lastname, is_male, birthday, height, description, location_id, goal, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING profile_id;
`

func (pr *ProfileRepo) StoreProfile(profile model.Profile) (profileId int, err error) {
	ctx := context.Background()

	var locationID *int
	if profile.Location != "" {
		parts := strings.Split(profile.Location, "@")
		if len(parts) != 3 {
			return 0, fmt.Errorf("invalid location format: expected 'Country@City@District'")
		}
		country := strings.TrimSpace(parts[0])
		city := strings.TrimSpace(parts[1])
		district := strings.TrimSpace(parts[2])

		var id int
		err := pr.DB.QueryRow(ctx, GetLocationID, country, city, district).Scan(&id)
		if err != nil {
			err = pr.DB.QueryRow(ctx, InsertLocation, country, city, district).Scan(&id)
			if err != nil {
				return 0, fmt.Errorf("failed to insert/get location: %w", err)
			}
		}
		locationID = &id
	}

	err = pr.DB.QueryRow(
		ctx,
		CreateProfileQuery,
		profile.FirstName,
		profile.LastName,
		profile.IsMale,
		profile.Birthday,
		profile.Height,
		profile.Description,
		locationID,
		profile.Goal,
	).Scan(&profileId)

	return
}

const GetProfilesQuery = `
WITH filtered_profiles AS (
    SELECT p.profile_id
    FROM profiles p
    LEFT JOIN likes liked 
        ON liked.liked_profile_id = p.profile_id AND liked.profile_id = $1
    JOIN users u ON u.profile_id = p.profile_id
    WHERE p.profile_id != $1
      AND liked.profile_id IS NULL
      AND u.user_id NOT IN (SELECT user_id FROM blacklist)
      AND ($2 = 0 OR p.profile_id > $2)
    ORDER BY p.profile_id
    LIMIT $3
)
SELECT DISTINCT ON (p.profile_id)
    p.profile_id, 
    p.firstname, 
    p.lastname, 
    p.is_male,
    p.height,
    p.birthday, 
    p.description,
    p.goal,
    l.country, 
    l.city,
    l.district,
    s.path AS avatar,
    i.description AS interest,
    pr.preference_description,
    pr.preference_value,
    param.parameter_description,
    param.parameter_value,
    CASE WHEN sbs.sub_id IS NOT NULL THEN TRUE ELSE FALSE END AS premium_status,
    sbs.border
FROM filtered_profiles fp
JOIN profiles p ON p.profile_id = fp.profile_id
LEFT JOIN locations l ON p.location_id = l.location_id
LEFT JOIN "static" s ON p.profile_id = s.profile_id
LEFT JOIN profile_interests pi ON pi.profile_id = p.profile_id
LEFT JOIN interests i ON pi.interest_id = i.interest_id
LEFT JOIN profile_preferences pp ON pp.profile_id = p.profile_id
LEFT JOIN preferences pr ON pp.preference_id = pr.preference_id
LEFT JOIN profile_parameter pp2 ON pp2.profile_id = p.profile_id
LEFT JOIN parameters param ON param.parameter_id = pp2.parameter_id
LEFT JOIN subscriptions sbs ON sbs.user_id = p.profile_id AND sbs.expires_at > NOW()
ORDER BY p.profile_id, i.description NULLS LAST;

`

func (pr *ProfileRepo) GetProfilesByUserId(forUserId int) ([]model.Profile, error) {
	const redisKeyFormat = "profiles_for_user:%d"
	redisKey := fmt.Sprintf(redisKeyFormat, forUserId)

	var lastSeenID int
	ctx := context.Background()

	result, err := pr.Client.Get(ctx, redisKey).Result()
	lastSeenID = 0
	if err == nil {
		if id, convErr := strconv.Atoi(result); convErr == nil {
			lastSeenID = id
		}
	} else if err != redis.Nil {
		return nil, err
	}

	rows, err := pr.DB.Query(ctx, GetProfilesQuery, forUserId, lastSeenID, model.PageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profileMap := make(map[int]*model.Profile)

	var maxProfileID int

	for rows.Next() {
		var (
			profileId                     int
			firstName, lastName           string
			isMale                        bool
			height                        int
			goal                          int
			birth                         sql.NullTime
			description                   sql.NullString
			country, city                 sql.NullString
			district                      sql.NullString
			photo                         sql.NullString
			interest                      sql.NullString
			preferenceDesc                sql.NullString
			preferenceValue               sql.NullString
			parameterDesc, parameterValue sql.NullString
			premiumStatus                 sql.NullBool
			premiumBorder                 sql.NullInt64
		)

		if err := rows.Scan(
			&profileId,
			&firstName,
			&lastName,
			&isMale,
			&height,
			&birth,
			&description,
			&goal,
			&country,
			&city,
			&district,
			&photo,
			&interest,
			&preferenceDesc,
			&preferenceValue,
			&parameterDesc,
			&parameterValue,
			&premiumStatus,
			&premiumBorder,
		); err != nil {
			return nil, err
		}

		if profileId > maxProfileID {
			maxProfileID = profileId
		}

		profile, exists := profileMap[profileId]
		if !exists {
			profile = &model.Profile{
				ProfileId:   profileId,
				FirstName:   firstName,
				LastName:    lastName,
				IsMale:      isMale,
				Goal:        goal,
				Height:      height,
				Interests:   []string{},
				Preferences: []model.Preference{},
				Photos:      []string{},
			}

			if birth.Valid {
				profile.Birthday = birth.Time
			}

			if description.Valid {
				profile.Description = description.String
			}

			if country.Valid && city.Valid && district.Valid {
				profile.Location = fmt.Sprintf("%s@%s@%s", country.String, city.String, district.String)
			}

			profileMap[profileId] = profile
		}

		if interest.Valid && !slices.Contains(profile.Interests, interest.String) {
			profile.Interests = append(profile.Interests, interest.String)
		}

		if parameterDesc.Valid && parameterValue.Valid {
			param := model.Preference{
				Description: parameterDesc.String,
				Value:       parameterValue.String,
			}
			if !slices.Contains(profile.Parameters, param) {
				profile.Parameters = append(profile.Parameters, param)
			}
		}

		if preferenceDesc.Valid && preferenceValue.Valid {
			pref := model.Preference{
				Description: preferenceDesc.String,
				Value:       preferenceValue.String,
			}
			if !slices.Contains(profile.Preferences, pref) {
				profile.Preferences = append(profile.Preferences, pref)
			}
		}

		if photo.Valid && !slices.Contains(profile.Photos, photo.String) {
			profile.Photos = append(profile.Photos, photo.String)
		}

		if premiumStatus.Valid && premiumStatus.Bool {
			profile.Premium.Status = true
			if premiumBorder.Valid {
				profile.Premium.Border = int(premiumBorder.Int64)
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	profiles := make([]model.Profile, 0, len(profileMap))
	for _, p := range profileMap {
		profiles = append(profiles, *p)
	}

	if maxProfileID > 0 {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = pr.Client.Set(ctx, redisKey, strconv.Itoa(maxProfileID), 10*time.Minute).Err()
		}()
	}

	return profiles, nil
}

const GetMatches = `
SELECT 
    profile_id, 
    matched_profile_id
FROM matches
WHERE profile_id = $1 OR matched_profile_id = $1;
`

func (pr *ProfileRepo) GetMatches(forUserId int) ([]model.Profile, error) {
	rows, err := pr.DB.Query(context.Background(), GetMatches, forUserId)
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
		profile, err := pr.GetProfileById(targ)
		if err != nil {
			return profiles, err
		}
		profiles = append(profiles, profile)
		amount++
	}
	return profiles, nil
}

const (
	UpdateProfileQuery = `
UPDATE profiles
SET
	firstname = $1,
	lastname = $2,
	is_male = $3,
	height = $4,
	description = $5,
	location_id = $6,
	birthday = $7,    
	goal = $8,
	updated_at = CURRENT_TIMESTAMP
WHERE profile_id = $9;

`

	DeleteProfileInterests = `
DELETE FROM profile_interests WHERE profile_id = $1
`

	DeleteProfilePreferences = `
DELETE FROM profile_preferences WHERE profile_id = $1
`

	GetPreferenceIDByFields = `
SELECT preference_id FROM preferences
WHERE preference_type = $1 AND preference_description = $2 AND preference_value = $3
`

	InsertPreferenceIfNotExists = `
INSERT INTO preferences (preference_type, preference_description, preference_value)
VALUES ($1, $2, $3)
RETURNING preference_id
`

	InsertProfilePreference = `
INSERT INTO profile_preferences (profile_id, preference_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING
`
	GetInterestIdByDescription = `
SELECT interest_id FROM interests WHERE description = $1
`
	InsertProfileInterest = `
INSERT INTO profile_interests (profile_id, interest_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING
`

	InsertStaticPhoto = `
INSERT INTO static (profile_id, path)
VALUES ($1, $2)
`
	InsertInterestIfNotExists = `
INSERT INTO interests (description)
VALUES ($1)
RETURNING interest_id
`
	GetLocationID = `
SELECT location_id FROM locations
WHERE country = $1 AND city = $2 AND district = $3
`

	InsertLocation = `
INSERT INTO locations (country, city, district)
VALUES ($1, $2, $3)
RETURNING location_id			
`

	DeleteProfileParameters = `
DELETE FROM profile_parameter WHERE profile_id = $1
`

	GetParameterIDByFields = `
SELECT parameter_id FROM parameters
WHERE parameter_type = $1 AND parameter_description = $2 AND parameter_value = $3
`

	InsertParameterIfNotExists = `
INSERT INTO parameters (parameter_type, parameter_description, parameter_value)
VALUES ($1, $2, $3)
RETURNING parameter_id
`

	InsertProfileParameter = `
INSERT INTO profile_parameter (profile_id, parameter_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING
`
)

func (pr *ProfileRepo) UpdateProfile(profileID int, newProfile model.Profile) error {
	ctx := context.Background()

	tx, err := pr.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var locationID int
	if newProfile.Location != "" {
		parts := strings.Split(newProfile.Location, "@")
		if len(parts) != 3 {
			return fmt.Errorf("invalid location format: expected 'Country@City@District'")
		}
		country, city, district := parts[0], parts[1], parts[2]

		err := tx.QueryRow(ctx, GetLocationID, country, city, district).Scan(&locationID)
		if err != nil {
			err = tx.QueryRow(ctx, InsertLocation, country, city, district).Scan(&locationID)
			if err != nil {
				return fmt.Errorf("failed to insert or get location: %w", err)
			}
		}
	}

	_, err = tx.Exec(ctx,
		UpdateProfileQuery,
		newProfile.FirstName,
		newProfile.LastName,
		newProfile.IsMale,
		newProfile.Height,
		newProfile.Description,
		locationID,
		newProfile.Birthday,
		newProfile.Goal,
		profileID,
	)
	if err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	if len(newProfile.Interests) != 0 {
		if _, err := tx.Exec(ctx, DeleteProfileInterests, profileID); err != nil {
			return fmt.Errorf("failed to delete old interests: %w", err)
		}

		for _, desc := range newProfile.Interests {
			var interestID int

			err := tx.QueryRow(ctx, GetInterestIdByDescription, desc).Scan(&interestID)
			if err != nil {
				err = tx.QueryRow(ctx, InsertInterestIfNotExists, desc).Scan(&interestID)
				if err != nil {
					return fmt.Errorf("failed to insert new interest '%s': %w", desc, err)
				}
			}

			_, err = tx.Exec(ctx, InsertProfileInterest, profileID, interestID)
			if err != nil {
				return fmt.Errorf("failed to insert profile interest: %w", err)
			}
		}
	}

	if len(newProfile.Preferences) != 0 {
		if _, err := tx.Exec(ctx, DeleteProfilePreferences, profileID); err != nil {
			return fmt.Errorf("failed to delete old preferences: %w", err)
		}

		for _, pref := range newProfile.Preferences {
			var preferenceID int

			err := tx.QueryRow(ctx, GetPreferenceIDByFields, 1, pref.Description, pref.Value).Scan(&preferenceID)
			if err != nil {
				err = tx.QueryRow(ctx, InsertPreferenceIfNotExists, 1, pref.Description, pref.Value).Scan(&preferenceID)
				if err != nil {
					return fmt.Errorf("failed to insert preference %+v: %w", pref, err)
				}
			}

			_, err = tx.Exec(ctx, InsertProfilePreference, profileID, preferenceID)
			if err != nil {
				return fmt.Errorf("failed to insert profile preference: %w", err)
			}
		}
	}

	if len(newProfile.Parameters) != 0 {
		if _, err := tx.Exec(ctx, DeleteProfileParameters, profileID); err != nil {
			return fmt.Errorf("failed to delete old parameters: %w", err)
		}

		for _, pref := range newProfile.Parameters {
			var parameterID int

			err := tx.QueryRow(ctx, GetParameterIDByFields, 1, pref.Description, pref.Value).Scan(&parameterID)
			if err != nil {
				err = tx.QueryRow(ctx, InsertParameterIfNotExists, 1, pref.Description, pref.Value).Scan(&parameterID)
				if err != nil {
					return fmt.Errorf("failed to insert parameter %+v: %w", pref, err)
				}
			}

			_, err = tx.Exec(ctx, InsertProfileParameter, profileID, parameterID)
			if err != nil {
				return fmt.Errorf("failed to insert profile parameter: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

const GetPhotoPathsQuery = `
SELECT path FROM static 
WHERE profile_id = (
	SELECT profile_id FROM users WHERE user_id = $1
);
`

func (pr *ProfileRepo) GetPhotos(userID int) ([]string, error) {
	rows, err := pr.DB.Query(context.Background(), GetPhotoPathsQuery, userID)
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

const (
	DeleteStaticQuery = `
	DELETE FROM "static" WHERE profile_id = $1 AND path = $2;
`
)

func (pr *ProfileRepo) DeletePhoto(profileID int, url string) error {
	cmdTag, err := pr.DB.Exec(context.Background(), DeleteStaticQuery, profileID, "/"+url)
	if err != nil {
		return fmt.Errorf("error deleting photo: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no photo found to delete")
	}

	return nil
}

const UploadPhotoQuery = `
INSERT INTO static (profile_id, path, created_at, updated_at)
VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING profile_id, path, created_at;
`

func (pr *ProfileRepo) StorePhoto(userID int, url string) error {
	_, err := pr.DB.Exec(context.Background(), UploadPhotoQuery, userID, url)
	return err
}

func (pr *ProfileRepo) CloseRepo() {
	if pool, ok := pr.DB.(*pgxpool.Pool); ok {
		pool.Close()
	}
}

const (
	CheckLikeExistsQuery = `
	SELECT like_id, status FROM likes
	WHERE profile_id = $1 AND liked_profile_id = $2 ;
	`

	CreateLikeQuery = `
	INSERT INTO likes (profile_id, liked_profile_id, created_at, status)
	VALUES ($1, $2, CURRENT_TIMESTAMP, $3)
	ON CONFLICT (profile_id, liked_profile_id)
	DO UPDATE SET
    	status = EXCLUDED.status,
    	created_at = CURRENT_TIMESTAMP
	RETURNING like_id;

	`

	CreateMatchQuery = `
	INSERT INTO matches (profile_id, matched_profile_id, created_at)
	VALUES ($1, $2, CURRENT_TIMESTAMP)
	`
	DeleteMatchQuery = `DELETE FROM matches WHERE 
				(profile_id = $1 AND matched_profile_id = $2) OR 
				(profile_id = $2 AND matched_profile_id = $1)`
)

func (pr *ProfileRepo) SetLike(from int, to int, status int) (likeID int, err error) {
	var existingID int
	var existing_status int
	err = pr.DB.QueryRow(
		context.Background(),
		CreateLikeQuery,
		from,
		to,
		status,
	).Scan(&likeID)

	if err != nil {
		return 0, fmt.Errorf("error inserting like: %w", err)
	}

	if status == 3 {
		err = pr.DB.QueryRow(context.Background(), CheckLikeExistsQuery, to, from).Scan(&existingID, &existing_status)
		if err == pgx.ErrNoRows {
			_, err = pr.DB.Exec(
				context.Background(),
				CreateLikeQuery,
				to,
				from,
				1,
			)
			if err != nil {
				return likeID, fmt.Errorf("error inserting reverse like: %w", err)
			}
		} else if err != nil {
			return likeID, fmt.Errorf("error checking reverse like: %w", err)
		}

		_, err = pr.DB.Exec(
			context.Background(),
			CreateMatchQuery,
			from,
			to,
		)
		if err != nil {
			return likeID, fmt.Errorf("error creating match: %w", err)
		}

		likeID = -1
		return likeID, nil
	}

	var reverseStatus int
	err = pr.DB.QueryRow(context.Background(), CheckLikeExistsQuery, to, from).Scan(&existingID, &reverseStatus)
	if err == nil && reverseStatus == 1 && status == 1 {
		_, err = pr.DB.Exec(
			context.Background(),
			CreateMatchQuery,
			from,
			to,
		)
		if err != nil {
			return likeID, fmt.Errorf("error creating match: %w", err)
		}
		likeID = -1
	}

	if status == 2 {
		_, err = pr.DB.Exec(
			context.Background(),
			DeleteMatchQuery,
			from, to,
		)
		if err != nil {
			return likeID, fmt.Errorf("error deleting match on dislike: %w", err)
		}
	}

	return likeID, nil
}

func (pr *ProfileRepo) StoreInterests(profileID int, interests []string) error {
	ctx := context.Background()

	tx, err := pr.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, desc := range interests {
		var interestID int

		err := tx.QueryRow(ctx, GetInterestIdByDescription, desc).Scan(&interestID)
		if err != nil {
			err = tx.QueryRow(ctx, InsertInterestIfNotExists, desc).Scan(&interestID)
			if err != nil {
				return err
			}
		}

		_, err = tx.Exec(ctx, InsertProfileInterest, profileID, interestID)
		if err != nil {
			return err
		}

	}

	return tx.Commit(ctx)
}

func (pr *ProfileRepo) StorePhotos(profileID int, paths []string) error {
	ctx := context.Background()

	for _, path := range paths {
		_, err := pr.DB.Exec(ctx, InsertStaticPhoto, profileID, path)
		if err != nil {
			return err
		}
	}
	return nil
}

const (
	DeleteProfileQuery = `
DELETE FROM profiles WHERE profile_id = $1;
`
	FindUserProfileQuery = `
	SELECT profile_id FROM users WHERE user_id = $1;
	`
)

func (pr *ProfileRepo) DeleteProfile(userId int) error {
	var profileId int
	err := pr.DB.QueryRow(context.Background(), FindUserProfileQuery, userId).Scan(&profileId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return model.ErrProfileNotFound
		}
		return model.ErrInvalidProfile
	}

	_, err = pr.DB.Exec(context.Background(), DeleteProfileQuery, profileId)
	if err != nil {
		return model.ErrDeleteProfile
	}

	return nil
}

const SearchProfilesQuery = `
WITH filtered_profiles AS (
    SELECT DISTINCT p.profile_id, p.firstname, p.lastname, p.birthday, p.goal,
           s.path AS avatar
    FROM profiles p
    JOIN users u ON u.profile_id = p.profile_id
    LEFT JOIN "static" s ON s.profile_id = p.profile_id
    LEFT JOIN likes liked ON liked.liked_profile_id = p.profile_id AND liked.profile_id = $1
    WHERE p.profile_id != $1
      AND liked.profile_id IS NULL
      AND u.user_id NOT IN (SELECT user_id FROM blacklist)
      AND (
          $2 = '' OR $2 = 'Any' OR
          (p.is_male = CASE 
                        WHEN $2 = 'Male' THEN true
                        WHEN $2 = 'Female' THEN false
                        ELSE NULL
                      END)
      )
      AND (
          $3 = 0 OR DATE_PART('year', AGE(CURRENT_DATE, p.birthday)) >= $3
      )
      AND (
          $4 = 0 OR DATE_PART('year', AGE(CURRENT_DATE, p.birthday)) <= $4
      )
      AND (
          $5 = 0 OR p.height >= $5
      )
      AND (
          $6 = 0 OR p.height <= $6
      )
      AND (
          $7 = 0 OR p.goal = $7
      )
      AND (
          $8 = '' OR EXISTS (
              SELECT 1 FROM locations l 
              WHERE l.location_id = p.location_id 
                AND LOWER(TRIM(l.country)) = LOWER(TRIM($8))
          )
      )
      AND (
          $9 = '' OR EXISTS (
              SELECT 1 FROM locations l 
              WHERE l.location_id = p.location_id 
                AND LOWER(TRIM(l.city)) = LOWER(TRIM($9))
          )
      )
      AND (
          NOT EXISTS (
            SELECT 1
            FROM jsonb_array_elements($10) AS pref(elem)
            WHERE NOT EXISTS (
                SELECT 1
                FROM profile_parameter pp
                JOIN parameters pr ON pr.parameter_id = pp.parameter_id
                WHERE pp.profile_id = p.profile_id
                  AND pr.parameter_description = pref.elem->>'preference_description'
                  AND pr.parameter_value = pref.elem->>'preference_value'
            )
          )
          OR jsonb_array_length($10) = 0
      )
      AND (
          $11 = '' OR (
              similarity((p.firstname || ' ' || p.lastname), $11) > 0.3
              OR similarity(p.fullname_translit, $11) > 0.3
              OR to_tsvector('russian', (p.firstname || ' ' || p.lastname)) @@ plainto_tsquery('russian', $11)
              OR to_tsvector('english', (p.firstname || ' ' || p.lastname)) @@ plainto_tsquery('english', $11)
              OR LOWER(p.firstname) LIKE LOWER($11 || '%')
              OR LOWER(p.lastname) LIKE LOWER($11 || '%')
          )
      )
)
SELECT DISTINCT ON (profile_id)
    profile_id AS "IDUser",
    avatar AS "FirstImg",
    firstname || ' ' || lastname AS "Fullname",
    FLOOR(DATE_PART('year', AGE(CURRENT_DATE, birthday)))::int AS "Age",
    goal AS "Goal"
FROM filtered_profiles
ORDER BY profile_id
LIMIT 20;


`

func (pr *ProfileRepo) SearchProfiles(cur_user int, params model.SearchProfileRequest) ([]model.FoundProfile, error) {
	ctx := context.Background()

	rows, err := pr.DB.Query(ctx, SearchProfilesQuery,
		cur_user,
		params.IsMale,
		params.AgeMin,
		params.AgeMax,
		params.HeightMin,
		params.HeightMax,
		params.Goal,
		params.Country,
		params.City,
		params.Preferences,
		params.Input,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.FoundProfile

	for rows.Next() {
		var fp model.FoundProfile
		if err := rows.Scan(
			&fp.IDUser,
			&fp.FirstImg,
			&fp.Fullname,
			&fp.Age,
			&fp.Goal,
		); err != nil {
			return nil, err
		}
		results = append(results, fp)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

const GetStaticticsQuery = `
SELECT 
	COUNT(DISTINCT l.like_id) FILTER (WHERE l.profile_id = p.profile_id) AS likes_given,
	COUNT(DISTINCT l2.like_id) FILTER (WHERE l2.liked_profile_id = p.profile_id) AS likes_received,
	COUNT(DISTINCT m.match_id) AS matches,
	COUNT(DISTINCT c1.complaint_id) AS complaints_made,
	COUNT(DISTINCT c2.complaint_id) AS complaints_received,
	COUNT(DISTINCT msg.message_id) AS messages_sent,
	COUNT(DISTINCT ch.chat_id) AS chat_count
FROM profiles p
LEFT JOIN users u ON u.profile_id = p.profile_id
LEFT JOIN likes l ON l.profile_id = p.profile_id
LEFT JOIN likes l2 ON l2.liked_profile_id = p.profile_id
LEFT JOIN matches m ON m.profile_id = p.profile_id OR m.matched_profile_id = p.profile_id
LEFT JOIN complaints c1 ON c1.complaint_by = u.user_id
LEFT JOIN complaints c2 ON c2.complaint_on = u.user_id
LEFT JOIN messages msg ON msg.user_id = p.profile_id
LEFT JOIN chats ch ON ch.first_profile_id = p.profile_id OR ch.second_profile_id = p.profile_id
WHERE p.profile_id = $1;
`

func (r *ProfileRepo) GetProfileStats(profileID int) (model.ProfileStats, error) {
	var stats model.ProfileStats
	err := r.DB.QueryRow(context.Background(), GetStaticticsQuery, profileID).Scan(
		&stats.LikesGiven,
		&stats.LikesReceived,
		&stats.Matches,
		&stats.ComplaintsMade,
		&stats.ComplaintsReceived,
		&stats.MessagesSent,
		&stats.ChatCount,
	)
	if err != nil {
		return stats, fmt.Errorf("failed to get profile stats: %w", err)
	}

	return stats, nil
}

const GetRecommendationsQuery = `
SELECT 
    bp.profile_id,
    bp.firstname,
    bp.lastname,
    bp.is_male,
    bp.height,
    bp.birthday,
    bp.description,
    bp.goal,
    l.country,
    l.city,
    l.district,
    liked.profile_id AS liked_by_profile_id,
    s.path AS avatar,
    i.description AS interest,
    pr.preference_description,
    pr.preference_value,
    param.parameter_description,
    param.parameter_value,
    CASE WHEN sbs.sub_id IS NOT NULL THEN TRUE ELSE FALSE END AS premium_status,
    sbs.border
FROM profiles bp
JOIN users bu ON bu.profile_id = bp.profile_id
LEFT JOIN profile_interests pi ON pi.profile_id = bp.profile_id
LEFT JOIN interests i ON i.interest_id = pi.interest_id
LEFT JOIN profile_preferences pp ON pp.profile_id = bp.profile_id
LEFT JOIN preferences pr ON pr.preference_id = pp.preference_id
LEFT JOIN profile_parameter ppar ON ppar.profile_id = bp.profile_id
LEFT JOIN parameters param ON param.parameter_id = ppar.parameter_id
LEFT JOIN locations l ON l.location_id = bp.location_id
LEFT JOIN static s ON bp.profile_id = s.profile_id
LEFT JOIN likes liked ON liked.liked_profile_id = bp.profile_id
LEFT JOIN subscriptions sbs ON sbs.user_id = bp.profile_id
LEFT JOIN blacklist bl ON bl.user_id = bu.user_id
WHERE 
    bl.user_id IS NULL 
    AND bp.profile_id != $1 
    AND NOT EXISTS (
        SELECT 1 FROM likes l2
        WHERE l2.profile_id = $1 AND l2.liked_profile_id = bp.profile_id
    )
    AND (
        SELECT COUNT(*) FROM profile_interests pi1
        WHERE pi1.profile_id = $1 AND pi1.interest_id IN (
            SELECT pi2.interest_id FROM profile_interests pi2 WHERE pi2.profile_id = bp.profile_id
        )
    ) * 1.0 / NULLIF((
        SELECT COUNT(*) FROM profile_interests pi3 WHERE pi3.profile_id = $1
    ), 0) >= 0.7
    AND (
        (
            SELECT COUNT(*) FROM profile_preferences  pp1
            WHERE pp1.profile_id = $1 AND pp1.preference_id IN (
                SELECT pp2.preference_id FROM profile_preferences pp2 WHERE pp2.profile_id = bp.profile_id
            )
        ) * 1.0 / NULLIF((
            SELECT COUNT(*) FROM profile_preferences pp3 WHERE pp3.profile_id = $1
        ), 0)
        +
        (
            SELECT COUNT(*) FROM profile_preferences pp4
            WHERE pp4.profile_id = bp.profile_id AND pp4.preference_id IN (
                SELECT pp5.preference_id FROM profile_preferences pp5 WHERE pp5.profile_id = $1
            )
        ) * 1.0 / NULLIF((
            SELECT COUNT(*) FROM profile_preferences pp6 WHERE pp6.profile_id = bp.profile_id
        ), 0)
    ) / 2.0 >= 0.5
LIMIT 1;
`

func (pr *ProfileRepo) GetRecomendations(profileId int) (model.Profile, error) {
	var profile model.Profile
	var birth sql.NullTime
	var interest sql.NullString
	var preferenceDesc sql.NullString
	var preferenceValue sql.NullString
	var likedByProfileId sql.NullInt64
	var photo sql.NullString
	var country, city, district sql.NullString

	var goal sql.NullInt64
	var paramDesc, paramValue sql.NullString

	var premiumStatus sql.NullBool
	var premiumBorder sql.NullInt64

	ctx := context.Background()
	rows, err := pr.DB.Query(ctx, GetRecommendationsQuery, profileId)
	fmt.Println(err)

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
			&goal,
			&country,
			&city,
			&district,
			&likedByProfileId,
			&photo,
			&interest,
			&preferenceDesc,
			&preferenceValue,
			&paramDesc,
			&paramValue,
			&premiumStatus,
			&premiumBorder,
		); err != nil {
			return profile, err
		}
		fmt.Println(err)

		if birth.Valid {
			profile.Birthday = birth.Time
		}

		if country.Valid && city.Valid && district.Valid {
			profile.Location = fmt.Sprintf("%s@%s@%s", country.String, city.String, district.String)
		}

		if likedByProfileId.Valid && !slices.Contains(profile.LikedBy, int(likedByProfileId.Int64)) {
			profile.LikedBy = append(profile.LikedBy, int(likedByProfileId.Int64))
		}

		if interest.Valid && !slices.Contains(profile.Interests, interest.String) {
			profile.Interests = append(profile.Interests, interest.String)
		}
		if goal.Valid {
			profile.Goal = int(goal.Int64)
		}

		if paramDesc.Valid && paramValue.Valid {
			param := model.Preference{
				Description: paramDesc.String,
				Value:       paramValue.String,
			}
			if !slices.Contains(profile.Parameters, param) {
				profile.Parameters = append(profile.Parameters, param)
			}
		}

		if preferenceDesc.Valid && preferenceValue.Valid {
			pref := model.Preference{
				Description: preferenceDesc.String,
				Value:       preferenceValue.String,
			}
			if !slices.Contains(profile.Preferences, pref) {
				profile.Preferences = append(profile.Preferences, pref)
			}
		}
		if photo.Valid && !slices.Contains(profile.Photos, photo.String) {
			profile.Photos = append(profile.Photos, photo.String)
		}
		if premiumStatus.Valid && premiumStatus.Bool {
			profile.Premium.Status = true
			if premiumBorder.Valid {
				profile.Premium.Border = int(premiumBorder.Int64)
			}
		}
	}

	if rows.Err() != nil {
		return profile, rows.Err()
	}

	return profile, nil
}
