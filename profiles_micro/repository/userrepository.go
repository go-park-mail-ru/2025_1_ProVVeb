package repository

import (
	"context"
	"database/sql"
	"fmt"
	"slices"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/model"
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
	CloseRepo() error
}

type ProfileRepo struct {
	DB *sql.DB
}

func NewUserRepo() (*ProfileRepo, error) {
	cfg := InitPostgresConfig()
	db, err := InitPostgresConnection(cfg)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return &ProfileRepo{}, err
	}
	return &ProfileRepo{DB: db}, nil
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

const GetProfileByIdQuery = `
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
    pr.preference_description,
	pr.preference_value 
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

func (pr *ProfileRepo) GetProfileById(profileId int) (model.Profile, error) {
	var profile model.Profile
	var birth sql.NullTime
	var interest sql.NullString
	var preferenceDesc sql.NullString
	var preferenceValue sql.NullString
	var likedByProfileId sql.NullInt64
	var photo sql.NullString
	var location sql.NullString

	rows, err := pr.DB.QueryContext(context.Background(), GetProfileByIdQuery, profileId)
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
			&preferenceDesc,
			&preferenceValue,
		); err != nil {
			return profile, err
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
	}

	if rows.Err() != nil {
		return profile, rows.Err()
	}

	return profile, nil
}

const CreateProfileQuery = `
INSERT INTO profiles (firstname, lastname, is_male, birthday, height, description, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING profile_id;
`

func (pr *ProfileRepo) StoreProfile(profile model.Profile) (profileId int, err error) {
	err = pr.DB.QueryRowContext(
		context.Background(),
		CreateProfileQuery,
		profile.FirstName,
		profile.LastName,
		profile.IsMale,
		profile.Birthday,
		profile.Height,
		profile.Description,
	).Scan(&profileId)
	return
}

func (pr *ProfileRepo) GetProfilesByUserId(forUserId int) ([]model.Profile, error) {
	profiles := make([]model.Profile, 0, model.PageSize)
	amount := 0
	for i := 1; ; i++ {
		if i != forUserId {
			profile, err := pr.GetProfileById(i)
			if err != nil {
				return profiles, err
			}
			if profile.ProfileId == 0 && profile.FirstName == "" {
				return profiles, nil
			}
			profiles = append(profiles, profile)
			amount++
		}
	}
}

const GetMatches = `
SELECT 
    profile_id, 
    matched_profile_id
FROM matches
WHERE profile_id = $1 OR matched_profile_id = $1;
`

func (pr *ProfileRepo) GetMatches(forUserId int) ([]model.Profile, error) {
	rows, err := pr.DB.QueryContext(context.Background(), GetMatches, forUserId)
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
	updated_at = CURRENT_TIMESTAMP
WHERE profile_id = $6;
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
	getInterestIdByDescription = `
SELECT interest_id FROM interests WHERE description = $1
`
	insertProfileInterest = `
INSERT INTO profile_interests (profile_id, interest_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING
`

	insertStaticPhoto = `
INSERT INTO static (profile_id, path)
VALUES ($1, $2)
`
	insertInterestIfNotExists = `
INSERT INTO interests (description)
VALUES ($1)
RETURNING interest_id
`
)

func (pr *ProfileRepo) UpdateProfile(profile_id int, new_profile model.Profile) error {
	ctx := context.Background()

	tx, err := pr.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(
		ctx,
		UpdateProfileQuery,
		new_profile.FirstName,
		new_profile.LastName,
		new_profile.IsMale,
		new_profile.Height,
		new_profile.Description,
		profile_id,
	)
	if err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	if len(new_profile.Interests) != 0 {

		if _, err := tx.ExecContext(ctx, DeleteProfileInterests, profile_id); err != nil {
			return fmt.Errorf("failed to delete old interests: %w", err)
		}

		for _, desc := range new_profile.Interests {
			var interestID int

			err := tx.QueryRowContext(ctx, getInterestIdByDescription, desc).Scan(&interestID)
			if err != nil {
				err = tx.QueryRowContext(ctx, insertInterestIfNotExists, desc).Scan(&interestID)
				if err != nil {
					return fmt.Errorf("failed to insert new interest '%s': %w", desc, err)
				}
			}

			_, err = tx.ExecContext(ctx, insertProfileInterest, profile_id, interestID)
			if err != nil {
				return fmt.Errorf("failed to insert profile interest: %w", err)
			}
		}
	}

	if len(new_profile.Preferences) != 0 {
		if _, err := tx.ExecContext(ctx, DeleteProfilePreferences, profile_id); err != nil {
			return fmt.Errorf("failed to delete old preferences: %w", err)
		}

		for _, pref := range new_profile.Preferences {
			var preferenceID int

			err := tx.QueryRowContext(ctx, GetPreferenceIDByFields,
				1, pref.Description, pref.Value,
			).Scan(&preferenceID)

			if err != nil {
				err = tx.QueryRowContext(ctx, InsertPreferenceIfNotExists,
					1, pref.Description, pref.Value,
				).Scan(&preferenceID)
				if err != nil {
					return fmt.Errorf("failed to insert preference %+v: %w", pref, err)
				}
			}

			_, err = tx.ExecContext(ctx, InsertProfilePreference, profile_id, preferenceID)
			if err != nil {
				return fmt.Errorf("failed to insert profile preference: %w", err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
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
	rows, err := pr.DB.QueryContext(context.Background(), GetPhotoPathsQuery, userID)
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
	result, err := pr.DB.ExecContext(context.Background(), DeleteStaticQuery, profileID, "/"+url)
	if err != nil {
		return fmt.Errorf("error deleting photo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
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
	_, err := pr.DB.ExecContext(context.Background(), UploadPhotoQuery, userID, url)
	return err
}

func (pr *ProfileRepo) CloseRepo() error {
	return ClosePostgresConnection(pr.DB)
}

const (
	CheckLikeExistsQuery = `
	SELECT like_id, status FROM likes
	WHERE profile_id = $1 AND liked_profile_id = $2;
	`

	CreateLikeQuery = `
	INSERT INTO likes (profile_id, liked_profile_id, created_at, status)
	VALUES ($1, $2, CURRENT_TIMESTAMP, $3)
	RETURNING like_id;
	`

	CreateMatchQuery = `
	INSERT INTO matches (profile_id, matched_profile_id, created_at)
	VALUES ($1, $2, CURRENT_TIMESTAMP)
	`
)

func (pr *ProfileRepo) SetLike(from int, to int, status int) (likeID int, err error) {
	var existingID int
	var existing_status int
	err = pr.DB.QueryRowContext(context.Background(), CheckLikeExistsQuery, from, to).Scan(&existingID, &existing_status)
	if err == nil {
		return 0, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("error checking existing like: %w", err)
	}
	err = pr.DB.QueryRowContext(
		context.Background(),
		CreateLikeQuery,
		from,
		to,
		status,
	).Scan(&likeID)

	if err != nil {
		return 0, fmt.Errorf("error inserting like: %w", err)
	}

	var reverseStatus int
	err = pr.DB.QueryRowContext(context.Background(), CheckLikeExistsQuery, to, from).Scan(&existingID, &reverseStatus)
	if err == nil && reverseStatus == 1 && status == 1 {
		_, err = pr.DB.ExecContext(
			context.Background(),
			CreateMatchQuery,
			from,
			to,
		)
		if err != nil {
			return likeID, fmt.Errorf("error creating match: %w", err)
		}

	}

	return likeID, nil
}

func (pr *ProfileRepo) StoreInterests(profileID int, interests []string) error {
	ctx := context.Background()

	tx, err := pr.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, desc := range interests {
		var interestID int

		err := tx.QueryRowContext(ctx, getInterestIdByDescription, desc).Scan(&interestID)
		if err != nil {
			err = tx.QueryRowContext(ctx, insertInterestIfNotExists, desc).Scan(&interestID)
			if err != nil {
				return err
			}
		}

		_, err = tx.ExecContext(ctx, insertProfileInterest, profileID, interestID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (pr *ProfileRepo) StorePhotos(profileID int, paths []string) error {
	ctx := context.Background()

	for _, path := range paths {
		_, err := pr.DB.ExecContext(ctx, insertStaticPhoto, profileID, path)
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
	fmt.Println("kjvnskkfn", userId)
	err := pr.DB.QueryRowContext(context.Background(), FindUserProfileQuery, userId).Scan(&profileId)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.ErrProfileNotFound
		}
		return model.ErrInvalidProfile
	}

	_, err = pr.DB.ExecContext(context.Background(), DeleteProfileQuery, profileId)
	if err != nil {
		return model.ErrDeleteProfile
	}

	return nil
}
