package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	querypb "github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/proto"
)

type GetAnswersForUser struct {
	QueryService querypb.QueryServiceClient
}

func NewGetAnswersForUserUseCase(queryService querypb.QueryServiceClient) (*GetAnswersForUser, error) {
	if queryService == nil {
		return nil, model.ErrGetActiveQueriesUC
	}
	return &GetAnswersForUser{QueryService: queryService}, nil
}

func (g *GetActiveQueries) GetAnswersForUser(userID int32) ([]model.QueryForUser, error) {
	req := &querypb.GetUserRequest{
		UserId: userID,
	}

	queryResp, err := g.QueryService.GetForUser(context.Background(), req)
	if err != nil {
		return nil, err
	}

	var queries []model.QueryForUser
	for _, q := range queryResp.Items {
		queries = append(queries, model.QueryForUser{
			Name:        q.Name,
			Description: q.Description,
			MinScore:    int(q.MinScore),
			MaxScore:    int(q.MaxScore),
			Score:       int(q.Score),
			Answer:      q.Answer,
		})
	}

	return queries, nil
}
