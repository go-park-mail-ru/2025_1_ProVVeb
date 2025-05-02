package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	querypb "github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/proto"
	"github.com/sirupsen/logrus"
)

type GetAnswersForUser struct {
	QueryService querypb.QueryServiceClient
	logger       *logger.LogrusLogger
}

func NewGetAnswersForUserUseCase(
	queryService querypb.QueryServiceClient,
	logger *logger.LogrusLogger,
) (*GetAnswersForUser, error) {
	if queryService == nil || logger == nil {
		return nil, model.ErrGetActiveQueriesUC
	}
	return &GetAnswersForUser{QueryService: queryService, logger: logger}, nil
}

func (g *GetAnswersForUser) GetAnswersForUser(userID int32) ([]model.QueryForUser, error) {
	g.logger.Info("GetAnswersForUser", "userID", userID)
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
	g.logger.WithFields(&logrus.Fields{"queriesCount": len(queries)}).Info("GetAnswersForUser")

	return queries, nil
}
