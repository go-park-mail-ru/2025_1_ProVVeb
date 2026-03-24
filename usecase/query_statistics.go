package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	querypb "github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/proto"
)

type GetStatistics struct {
	QueryService querypb.QueryServiceClient
	logger       *logger.LogrusLogger
}

func NewGetStatisticsUseCase(
	queryService querypb.QueryServiceClient,
	logger *logger.LogrusLogger,
) (*GetStatistics, error) {
	if queryService == nil || logger == nil {
		return nil, model.ErrGetActiveQueriesUC
	}
	return &GetStatistics{QueryService: queryService, logger: logger}, nil
}

func (g *GetStatistics) GetStatistics(query_name string) (model.QueryStats, error) {
	g.logger.Info("GetStatistics", "query_name", query_name)
	req := &querypb.QueryStatsRequest{
		QueryName: query_name,
	}
	var Stats model.QueryStats

	queryResp, err := g.QueryService.GetQueryStats(context.Background(), req)
	if err != nil {
		g.logger.Error("GetStatistics", "gRPC not working")
		return Stats, err
	}

	Stats.AverageScore = queryResp.AverageScore
	Stats.TotalAnswers = int(queryResp.TotalAnswers)
	Stats.MinScore = int(queryResp.MinScore)
	Stats.MaxScore = int(queryResp.MaxScore)

	return Stats, nil
}
