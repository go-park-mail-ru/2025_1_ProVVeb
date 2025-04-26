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

func (g *GetActiveQueries) GetActiveQueries(userID uint32) error{
	// return g.QueryService.GetActiveQueries(userID)
	req := g.QueryService.GetActive(context.Background(), &querypb.GetUserRequest{UserId: userID})

}
