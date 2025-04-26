package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	querypb "github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/proto"
)

type GetActiveQueries struct {
	QueryService querypb.QueryServiceClient
}

func NewGetActiveQueriesUseCase(queryService querypb.QueryServiceClient) (*GetActiveQueries, error) {
	if queryService == nil {
		return nil, model.ErrGetActiveQueriesUC
	}
	return &GetActiveQueries{QueryService: queryService}, nil
}

func (g *GetActiveQueries) GetActiveQueries(userID int32) ([]model.Query, error) {
	req := &querypb.GetUserRequest{
		UserId: userID,
	}

	queryResp, err := g.QueryService.GetActive(context.Background(), req)
	if err != nil {
		return nil, err
	}

	var queries []model.Query
	for _, q := range queryResp.Items {
		queries = append(queries, model.Query{
			Name:        q.Name,
			Description: q.Description,
			MinScore:    int(q.MinScore),
			MaxScore:    int(q.MaxScore),
		})
	}

	return queries, nil
}
