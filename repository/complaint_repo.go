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
