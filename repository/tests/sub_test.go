package repository

import (
	"context"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/stretchr/testify/assert"
)

func TestSubRepo_CreateSub(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sr := &repository.SubRepo{
		DB:  db,
		Ctx: context.Background(),
	}

	// Ожидаемый SQL и аргументы
	mock.ExpectExec(regexp.QuoteMeta(repository.CreateSubQuery)).
		WithArgs(1, 2, "data").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = sr.CreateSub(1, 2, "data")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSubRepo_UpdateBorder(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sr := &repository.SubRepo{
		DB:  db,
		Ctx: context.Background(),
	}

	mock.ExpectExec(regexp.QuoteMeta(repository.UpdateBorderQuery)).
		WithArgs(100, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = sr.UpdateBorder(1, 100)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
