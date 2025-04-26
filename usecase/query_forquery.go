package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	querypb "github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GetAnswersForQuery struct {
	QueryService querypb.QueryServiceClient
}

func NewGetAnswersForQueryUseCase(queryService querypb.QueryServiceClient) (*GetAnswersForQuery, error) {
	if queryService == nil {
		return nil, model.ErrGetActiveQueriesUC
	}
	return &GetAnswersForQuery{QueryService: queryService}, nil
}

func (g *GetActiveQueries) GetAnswersForQuery() ([]model.UsersForQuery, error) {
	req := &emptypb.Empty{}

	queryResp, err := g.QueryService.GetForQuery(context.Background(), req)
	if err != nil {
		return nil, err
	}

	var queries []model.UsersForQuery
	for _, q := range queryResp.Items {
		queries = append(queries, model.UsersForQuery{
			Name:        q.Name,
			Description: q.Description,
			MinScore:    int(q.MinScore),
			MaxScore:    int(q.MaxScore),
			Score:       int(q.Score),
			Answer:      q.Answer,
			Login:       q.Login,
		})
	}

	return queries, nil
}
