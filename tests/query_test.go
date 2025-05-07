package tests

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/config"
	query "github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/server"
)

func TestQueryRepo_GetActive(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &query.QueryRepo{DB: db}

	const userID = 42

	rows := sqlmock.NewRows([]string{"name", "description", "min_score", "max_score"}).
		AddRow("query1", "desc1", 1, 10).
		AddRow("query2", "desc2", 5, 15)

	mock.ExpectQuery("SELECT (.+) FROM queries").
		WithArgs(userID).
		WillReturnRows(rows)

	queries, err := repo.GetActive(userID)
	require.NoError(t, err)
	require.Len(t, queries, 2)
	require.Equal(t, "query1", queries[0].Name)
	require.Equal(t, 5, queries[1].MinScore)
}

func TestQueryRepo_SendResp(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &query.QueryRepo{DB: db}

	answer := config.Answer{
		QueryName: "query1",
		UserId:    42,
		Score:     7,
		Answer:    "yes",
	}

	mock.ExpectQuery("INSERT INTO user_answer").
		WithArgs(answer.QueryName, answer.UserId, answer.Score, answer.Answer).
		WillReturnRows(sqlmock.NewRows([]string{"answer_id"}).AddRow(1))

	err = repo.SendResp(answer)
	require.NoError(t, err)
}

func TestQueryRepo_GetForUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &query.QueryRepo{DB: db}

	const userID = 7

	rows := sqlmock.NewRows([]string{"name", "description", "min_score", "max_score", "score", "answer"}).
		AddRow("q1", "desc1", 0, 10, 8, "yes").
		AddRow("q2", "desc2", 5, 15, 6, "no")

	mock.ExpectQuery("SELECT (.+) FROM user_answer").
		WithArgs(userID).
		WillReturnRows(rows)

	answers, err := repo.GetForUser(userID)
	require.NoError(t, err)
	require.Len(t, answers, 2)
	require.Equal(t, "desc2", answers[1].Description)
	require.Equal(t, "no", answers[1].Answer)
}

func TestQueryRepo_GetUsersForQueries(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &query.QueryRepo{DB: db}

	rows := sqlmock.NewRows([]string{"name", "description", "min_score", "max_score", "login", "answer", "score"}).
		AddRow("q1", "desc1", 1, 10, "user1", "yes", 9).
		AddRow("q2", "desc2", 2, 20, "user2", "maybe", 15)

	mock.ExpectQuery("SELECT (.+) FROM user_answer").
		WillReturnRows(rows)

	users, err := repo.GetUsersForQueries()
	require.NoError(t, err)
	require.Len(t, users, 2)
	require.Equal(t, "user2", users[1].Login)
	require.Equal(t, 15, users[1].Score)
}
