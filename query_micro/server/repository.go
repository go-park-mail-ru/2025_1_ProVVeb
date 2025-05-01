package query

import (
	"context"

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
