package query

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/config"
)

const GetActiveQuery = `
SELECT 
    q.name,
    q.description,
    q.min_score,
    q.max_score
FROM queries q
LEFT JOIN user_answer ua 
    ON q.query_id = ua.query_id 
    AND ua.user_id = $1 
    AND ua.created_at <= NOW() - INTERVAL '7 days'
WHERE q.is_active = TRUE
  AND ua.answer_id IS NULL;
`

func (qr *QueryRepo) GetActive(forUserId int) ([]config.Query, error) {
	var query_array []config.Query
	rows, err := qr.DB.QueryContext(context.Background(), GetActiveQuery, forUserId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var query config.Query
		if err := rows.Scan(&query.Name, &query.Description, &query.MinScore, &query.MaxScore); err != nil {
			return nil, err
		}
		query_array = append(query_array, query)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return query_array, nil
}

const SendRespQuery = `
INSERT INTO user_answer (query_id, user_id, score, answer)
VALUES (
    (SELECT query_id FROM queries WHERE name = $1),
    $2,
    $3,
    $4
)
RETURNING answer_id;
`

func (qr *QueryRepo) SendResp(answer config.Answer) error {
	var respId int

	err := qr.DB.QueryRowContext(
		context.Background(),
		SendRespQuery,
		answer.QueryName,
		answer.UserId,
		answer.Score,
		answer.Answer,
	).Scan(&respId)

	return err
}

const GetForUserQuery = `
SELECT 
    q.name,
    q.description,
    q.min_score,
    q.max_score,
    ua.score,
    ua.answer
FROM user_answer ua
JOIN queries q ON ua.query_id = q.query_id
WHERE ua.user_id = $1;
`

func (qr *QueryRepo) GetForUser(forUserId int) ([]config.QueryForUser, error) {
	var answer_array []config.QueryForUser
	rows, err := qr.DB.QueryContext(context.Background(), GetForUserQuery, forUserId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var answer config.QueryForUser
		err := rows.Scan(
			&answer.Name,
			&answer.Description,
			&answer.MinScore,
			&answer.MaxScore,
			&answer.Score,
			&answer.Answer,
		)
		if err != nil {
			return nil, err
		}
		answer_array = append(answer_array, answer)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return answer_array, nil
}

const GetAllQueries = `
SELECT 
    q.name,
    q.description,
    q.min_score,
    q.max_score,
    u.login,
    ua.answer,
    ua.score
FROM user_answer ua
JOIN queries q ON ua.query_id = q.query_id
JOIN users u ON ua.user_id = u.user_id
`

func (qr *QueryRepo) GetUsersForQueries() ([]config.UsersForQuery, error) {
	var usersForQuery []config.UsersForQuery

	rows, err := qr.DB.QueryContext(context.Background(), GetAllQueries)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userForQuery config.UsersForQuery
		err := rows.Scan(
			&userForQuery.Name,
			&userForQuery.Description,
			&userForQuery.MinScore,
			&userForQuery.MaxScore,
			&userForQuery.Login,
			&userForQuery.Answer,
			&userForQuery.Score,
		)
		if err != nil {
			return nil, err
		}
		usersForQuery = append(usersForQuery, userForQuery)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return usersForQuery, nil
}

const FindQuery = `
SELECT 
    q.name,
    q.description,
    q.min_score,
    q.max_score,
    u.login,
    ua.answer,
    ua.score,
    ua.user_id
FROM user_answer ua
JOIN queries q ON ua.query_id = q.query_id
JOIN users u ON ua.user_id = u.user_id
JOIN profiles p ON p.profile_id = u.user_id
WHERE q.is_active = TRUE
  AND ($1::BIGINT = 0 OR q.query_id = $1)
  AND (
        $2 = '' OR
        similarity(u.login, $2) > 0.3 OR
        LOWER(u.login) LIKE LOWER($2 || '%')
  )

`

func (qr *QueryRepo) FindQuery(name string, queryID int) ([]config.AnswersForQuery, error) {
	rows, err := qr.DB.QueryContext(context.Background(), FindQuery, queryID, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []config.AnswersForQuery
	for rows.Next() {
		var row config.AnswersForQuery
		if err := rows.Scan(
			&row.Name,
			&row.Description,
			&row.MinScore,
			&row.MaxScore,
			&row.Login,
			&row.Answer,
			&row.Score,
			&row.UserId,
		); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, nil
}

const DeleteAnswerQuery = `
DELETE FROM user_answer
USING queries
WHERE user_answer.query_id = queries.query_id
  AND queries.name = $1
  AND user_answer.user_id = $2;
`

func (qr *QueryRepo) DeleteAnswer(query_name string, user_id int) error {
	_, err := qr.DB.ExecContext(context.Background(), DeleteAnswerQuery, query_name, user_id)
	if err != nil {
		return model.ErrDeleteUser
	}
	return nil
}

const getStatisticsQuery = `
SELECT 
    COUNT(*) AS total_answers,
    AVG(score) AS average_score,
    MIN(score) AS min_score,
    MAX(score) AS max_score
FROM user_answer ua
JOIN queries q ON ua.query_id = q.query_id
WHERE ($1 = '' OR q.name = $1);

`

func (qr *QueryRepo) GetStatistics(queryName string) (config.QueryStats, error) {
	var (
		totalAnswers int64
		avgScore     sql.NullFloat64
		minScore     sql.NullInt64
		maxScore     sql.NullInt64
	)

	row := qr.DB.QueryRowContext(context.Background(), getStatisticsQuery, queryName)
	err := row.Scan(&totalAnswers, &avgScore, &minScore, &maxScore)
	if err != nil {
		return config.QueryStats{}, fmt.Errorf("failed to get statistics: %w", err)
	}

	return config.QueryStats{
		TotalAnswers: totalAnswers,
		AverageScore: avgScore.Float64,
		MinScore:     int(minScore.Int64),
		MaxScore:     int(maxScore.Int64),
	}, nil
}
