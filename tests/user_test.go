package tests

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	profilesrepo "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/repository"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// func UserRepoTestSQL_StoreInterests(t *testing.T) {
// 	db, mock, err := sqlmock.New()
// 	require.NoError(t, err)
// 	defer db.Close()

// 	repo := &profilesrepo.ProfileRepo{DB: db}

// 	mock.ExpectBegin()

// 	mock.ExpectQuery(`SELECT interest_id FROM interests WHERE description = \$1`).
// 		WithArgs("Музыка").
// 		WillReturnError(sql.ErrNoRows)

// 	mock.ExpectQuery(`INSERT INTO interests \(description\) VALUES \(\$1\) RETURNING interest_id`).
// 		WithArgs("Музыка").
// 		WillReturnRows(sqlmock.NewRows([]string{"interest_id"}).AddRow(3))

// 	mock.ExpectExec(`INSERT INTO profile_interests \(profile_id, interest_id\) VALUES \(\$1, \$2\)`).
// 		WithArgs(1, 3).
// 		WillReturnResult(sqlmock.NewResult(1, 1))

// 	mock.ExpectCommit()

// 	err = repo.StoreInterests(1, []string{"Музыка"})
// 	require.NoError(t, err)
// 	require.NoError(t, mock.ExpectationsWereMet())
// }

func TestStoreInterests(t *testing.T) {
	mockDB := new(MockDB)
	repo := &profilesrepo.ProfileRepo{DB: mockDB}

	mockDB.On("Query", mock.Anything,
		`SELECT interest_id FROM interests WHERE description = $1`,
		[]interface{}{"Музыка"}).
		Return(&MockRows{}, sql.ErrNoRows)

	mockDB.On("Query", mock.Anything,
		`INSERT INTO interests (description) VALUES ($1) RETURNING interest_id`,
		[]interface{}{"Музыка"}).
		Return(&MockRows{
			data: [][]interface{}{{3}},
		}, nil)

	mockDB.On("Exec", mock.Anything,
		`INSERT INTO profile_interests (profile_id, interest_id) VALUES ($1, $2)`,
		[]interface{}{1, 3}).
		Return(pgconn.NewCommandTag("INSERT 1"), nil)

	mockDB.On("Commit", mock.Anything).Return(nil)

	err := repo.StoreInterests(1, []string{"Музыка"})
	require.NoError(t, err)

	mockDB.AssertExpectations(t)
}

func TestComplaintRepo_CreateComplaint(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &repository.ComplaintRepo{DB: db}

	const (
		complaintBy   = 1
		complaintOn   = 2
		complaintType = "Спам"
		text          = "Он мне пишет рекламу"
	)
	var complaintTypeID = 42

	mock.ExpectQuery("SELECT comp_type FROM complaint_types").
		WithArgs(complaintType).
		WillReturnRows(sqlmock.NewRows([]string{"comp_type"}).AddRow(complaintTypeID))

	mock.ExpectExec("INSERT INTO complaints").
		WithArgs(complaintBy, complaintOn, complaintTypeID, text, 1, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateComplaint(complaintBy, complaintOn, complaintType, text)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestComplaintRepo_CreateComplaint_TypeNotExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &repository.ComplaintRepo{DB: db}

	const (
		complaintBy   = 3
		complaintType = "Нарушение правил"
		text          = "Неприемлемое поведение"
	)
	var insertedTypeID = 99

	mock.ExpectQuery("SELECT comp_type FROM complaint_types").
		WithArgs(complaintType).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery("INSERT INTO complaint_types").
		WithArgs(complaintType).
		WillReturnRows(sqlmock.NewRows([]string{"comp_type"}).AddRow(insertedTypeID))

	mock.ExpectExec("INSERT INTO complaints").
		WithArgs(complaintBy, complaintBy, insertedTypeID, text, 1, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateComplaint(complaintBy, 0, complaintType, text)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
