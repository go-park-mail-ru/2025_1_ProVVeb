package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	querypb "github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GetAnswersForQuery struct {
	QueryService querypb.QueryServiceClient
	logger       *logger.LogrusLogger
}

func NewGetAnswersForQueryUseCase(
	queryService querypb.QueryServiceClient,
	logger *logger.LogrusLogger,
) (*GetAnswersForQuery, error) {
	if queryService == nil || logger == nil {
		return nil, model.ErrGetActiveQueriesUC
	}
	return &GetAnswersForQuery{QueryService: queryService, logger: logger}, nil
}

func (g *GetAnswersForQuery) GetAnswersForQuery() ([]model.UsersForQuery, error) {
	g.logger.Info("GetAnswersForQuery")
	req := &emptypb.Empty{}

	queryResp, err := g.QueryService.GetForQuery(context.Background(), req)
	if err != nil {
		g.logger.Error("GetAnswersForQuery", "error", err)
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

	g.logger.WithFields(&logrus.Fields{"queriesCount": len(queries)}).Info("GetAnswersForQuery")
	return queries, nil
}
