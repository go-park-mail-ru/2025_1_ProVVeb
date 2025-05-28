package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	querypb "github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/proto"
	"github.com/sirupsen/logrus"
)

type FindQuery struct {
	QueryService querypb.QueryServiceClient
	logger       *logger.LogrusLogger
}

func NewFindQueryUseCase(
	queryService querypb.QueryServiceClient,
	logger *logger.LogrusLogger,
) (*FindQuery, error) {
	if queryService == nil || logger == nil {
		return nil, model.ErrGetActiveQueriesUC
	}
	return &FindQuery{QueryService: queryService, logger: logger}, nil
}

func (g *FindQuery) FindQuery(Name string, query_id int) ([]model.AnswersForQuery, error) {
	g.logger.Info("GetAnswersForQuery")
	req := &querypb.FindQueryRequest{
		Name:    Name,
		QueryId: int32(query_id),
	}

	queryResp, err := g.QueryService.FindQuery(context.Background(), req)
	if err != nil {
		g.logger.Error("GetAnswersForQuery", "error", err)
		return nil, err
	}

	var queries []model.AnswersForQuery
	for _, q := range queryResp.Items {
		queries = append(queries, model.AnswersForQuery{
			Name:        q.Name,
			Description: q.Description,
			MinScore:    int(q.MinScore),
			MaxScore:    int(q.MaxScore),
			Score:       int(q.Score),
			Answer:      q.Answer,
			Login:       q.Login,
			UserId:      int(q.UserId),
		})
	}

	g.logger.WithFields(&logrus.Fields{"queriesCount": len(queries)}).Info("GetAnswersForQuery")
	return queries, nil
}
