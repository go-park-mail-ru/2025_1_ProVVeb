package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type ComplaintRepository interface {
	CreateComplaint(complaint_by int, complaint_on int, ComplaintType string, text string) error
	GetAllComplaints(ctx context.Context) ([]model.ComplaintWithLogins, error)
	FindComplaint(complaint_by int, name_by string, complaint_on int, name_on string, complaint_type string, status int) ([]model.ComplaintWithLogins, error)
	HandleComplaint(complaint_id int, new_status int) error
	GetStatistics(useFrom bool, from time.Time, useTo bool, to time.Time) (model.ComplaintStats, error)
	DeleteComplaint(complaint_id int) error
}

type ComplaintRepo struct {
	DB *sql.DB
}

func NewComplaintRepo() (*ComplaintRepo, error) {
	cfg := InitPostgresConfig()
	db, err := InitPostgresConnection(cfg)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return &ComplaintRepo{}, err
	}
	return &ComplaintRepo{
		DB: db,
	}, nil
}

const (
	GetComplaintTypeIDQuery = `
		SELECT comp_type FROM complaint_types WHERE type_description = $1
	`

	InsertComplaintTypeQuery = `
		INSERT INTO complaint_types (type_description)
		VALUES ($1)
		RETURNING comp_type
	`

	InsertComplaintQuery = `
		INSERT INTO complaints (
			complaint_by,
			complaint_on,
			complaint_type,
			complaint_text,
			status,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
)

func (r *ComplaintRepo) CreateComplaint(complaintBy int, complaintOn int, complaintType string, text string) error {
	ctx := context.Background()

	var compTypeID int
	err := r.DB.QueryRowContext(ctx, GetComplaintTypeIDQuery, complaintType).Scan(&compTypeID)
	if err == sql.ErrNoRows {
		err = r.DB.QueryRowContext(ctx, InsertComplaintTypeQuery, complaintType).Scan(&compTypeID)
		if err != nil {
			return fmt.Errorf("failed to insert new complaint_type: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to fetch complaint_type: %w", err)
	}
	if complaintOn == 0 {
		complaintOn = complaintBy
	}

	_, err = r.DB.ExecContext(ctx, InsertComplaintQuery,
		complaintBy,
		complaintOn,
		compTypeID,
		text,
		1,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to insert complaint: %w", err)
	}

	return nil
}

const GetAllComplaintsQuery = `
    SELECT 
        c.complaint_id,
        u_by.login AS complaint_by_login,
        u_on.login AS complaint_on_login,
        c.complaint_type,
        t.type_description,
        c.complaint_text,
        c.status,
        c.created_at,
        c.closed_at
    FROM complaints c
    LEFT JOIN users u_by ON c.complaint_by = u_by.user_id
    LEFT JOIN users u_on ON c.complaint_on = u_on.user_id
    JOIN complaint_types t ON c.complaint_type = t.comp_type
    ORDER BY c.created_at DESC
`

func (r *ComplaintRepo) GetAllComplaints(ctx context.Context) ([]model.ComplaintWithLogins, error) {
	rows, err := r.DB.QueryContext(ctx, GetAllComplaintsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var complaints []model.ComplaintWithLogins

	for rows.Next() {
		var c model.ComplaintWithLogins
		err := rows.Scan(
			&c.ComplaintID,
			&c.ComplaintBy,
			&c.ComplaintOn,
			&c.ComplaintType,
			&c.TypeDesc,
			&c.Text,
			&c.Status,
			&c.CreatedAt,
			&c.ClosedAt,
		)
		if err != nil {
			return nil, err
		}
		complaints = append(complaints, c)
	}

	return complaints, rows.Err()
}

const findComplaintsQuery = `
WITH filtered_complaints AS (
    SELECT
        c.complaint_id,
        cb.login AS complaint_by,
        co.login AS complaint_on,
        c.complaint_type,
        ct.type_description,
        c.complaint_text,
        c.status,
        c.created_at,
        c.closed_at
    FROM complaints c
    JOIN users cb ON cb.user_id = c.complaint_by
    JOIN users co ON co.user_id = c.complaint_on
    JOIN complaint_types ct ON ct.comp_type = c.complaint_type
    WHERE
        ($1 = 0 OR c.complaint_by = $1)
      AND ($2 = 0 OR c.complaint_on = $2)
      AND ($3 = '' OR LOWER(ct.type_description) = LOWER($3))
      AND ($4 = 0 OR c.status = $4)
      AND (
          ($5 = '' AND $6 = '')
          OR (
              $5 <> '' AND (
                  similarity(cb.login, $5) > 0.3
                  OR LOWER(cb.login) LIKE LOWER($5 || '%')
              )
          )
          OR (
              $6 <> '' AND (
                  similarity(co.login, $6) > 0.3
                  OR LOWER(co.login) LIKE LOWER($6 || '%')
              )
          )
      )
)
SELECT
    complaint_id,
    complaint_by,
    complaint_on,
    complaint_type,
    type_description,
    complaint_text,
    status,
    created_at,
    closed_at
FROM filtered_complaints
ORDER BY created_at DESC;


`

func (cr *ComplaintRepo) FindComplaint(
	complaintById int,
	nameBy string,
	complaintOnId int,
	nameOn string,
	complaintType string,
	status int,
) ([]model.ComplaintWithLogins, error) {
	var closedAt sql.NullTime
	rows, err := cr.DB.QueryContext(
		context.Background(),
		findComplaintsQuery,
		complaintById, complaintOnId, complaintType, status, nameBy, nameOn)
	fmt.Println(rows)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.ComplaintWithLogins
	for rows.Next() {
		var row model.ComplaintWithLogins
		if err := rows.Scan(
			&row.ComplaintID,
			&row.ComplaintBy,
			&row.ComplaintOn,
			&row.ComplaintType,
			&row.TypeDesc,
			&row.Text,
			&row.Status,
			&row.CreatedAt,
			&closedAt,
		); err != nil {
			return nil, err
		}
		if closedAt.Valid {
			row.ClosedAt = &closedAt.Time
		} else {
			row.ClosedAt = nil
		}

		result = append(result, row)
	}
	fmt.Println(result)
	return result, nil
}

const DeleteComplaintQuery = `
DELETE FROM complaints WHERE complaint_id = $1;
`

func (cr *ComplaintRepo) DeleteComplaint(complaint_id int) error {
	_, err := cr.DB.ExecContext(context.Background(), DeleteComplaintQuery, complaint_id)
	if err != nil {
		return model.ErrDeleteUser
	}
	return nil
}

const (
	GetTargetUserQuery = `
		SELECT complaint_on
		FROM complaints
		WHERE complaint_id = $1
	`

	UpdateComplaintQuery = `
		UPDATE complaints
		SET status = $1, closed_at = CURRENT_TIMESTAMP
		WHERE complaint_id = $2
	`

	BlockUserQuery = `
			INSERT INTO blacklist (user_id)
			SELECT $1
			WHERE NOT EXISTS (
				SELECT 1 FROM blacklist WHERE user_id = $1
			)
		`
)

func (cr *ComplaintRepo) HandleComplaint(complaint_id int, new_status int) error {
	ctx := context.Background()
	tx, err := cr.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var targetUserID int
	err = tx.QueryRowContext(ctx, GetTargetUserQuery, complaint_id).Scan(&targetUserID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, UpdateComplaintQuery, new_status, complaint_id)
	if err != nil {
		return err
	}

	if new_status == 2 {
		_, err = tx.ExecContext(ctx, BlockUserQuery, targetUserID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

const getStatisticsQuery = `
	SELECT
		COUNT(*) AS total_complaints,
		COUNT(*) FILTER (WHERE status = -1) AS rejected,
		COUNT(*) FILTER (WHERE status = 1) AS pending,
		COUNT(*) FILTER (WHERE status = 2) AS approved,
		COUNT(*) FILTER (WHERE status = 3) AS closed,
		COUNT(DISTINCT complaint_by) AS total_complainants,
		COUNT(DISTINCT complaint_on) AS total_reported,
		COALESCE(MIN(created_at), NOW()) AS first_complaint,
		COALESCE(MAX(created_at), NOW()) AS last_complaint
	FROM complaints
	WHERE
		($1::bool IS FALSE OR created_at >= $2)
	  AND ($3::bool IS FALSE OR created_at <= $4)
`

func (cr *ComplaintRepo) GetStatistics(useFrom bool, from time.Time, useTo bool, to time.Time) (model.ComplaintStats, error) {
	row := cr.DB.QueryRowContext(context.Background(), getStatisticsQuery, useFrom, from, useTo, to)

	var stats model.ComplaintStats
	err := row.Scan(
		&stats.Total,
		&stats.Rejected,
		&stats.Pending,
		&stats.Approved,
		&stats.Closed,
		&stats.TotalBy,
		&stats.TotalOn,
		&stats.FirstComplaint,
		&stats.LastComplaint,
	)
	if err != nil {
		return stats, err
	}

	return stats, nil
}
