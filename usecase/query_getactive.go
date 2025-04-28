package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	querypb "github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/proto"
	"github.com/sirupsen/logrus"
)

type GetActiveQueries struct {
	QueryService querypb.QueryServiceClient
	logger       *logger.LogrusLogger
}

func NewGetActiveQueriesUseCase(
	queryService querypb.QueryServiceClient,
	logger *logger.LogrusLogger,
) (*GetActiveQueries, error) {
	if queryService == nil || logger == nil {
		return nil, model.ErrGetActiveQueriesUC
	}
	return &GetActiveQueries{QueryService: queryService, logger: logger}, nil
}

func (g *GetActiveQueries) GetActiveQueries(userID int32) ([]model.Query, error) {
	g.logger.Info("GetActiveQueries", "userID", userID)
	req := &querypb.GetUserRequest{
		UserId: userID,
	}

	queryResp, err := g.QueryService.GetActive(context.Background(), req)
	if err != nil {
		g.logger.Error("GetActiveQueries", "gRPC not working")
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

	g.logger.WithFields(&logrus.Fields{"queriesCount": len(queries)})
	return queries, nil
}
