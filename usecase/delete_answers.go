package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	querypb "github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/proto"
)

type DeleteAnswer struct {
	QueryService querypb.QueryServiceClient
	logger       *logger.LogrusLogger
}

func NewDeleteAnswerUseCase(
	queryService querypb.QueryServiceClient,
	logger *logger.LogrusLogger,
) (*DeleteAnswer, error) {
	if queryService == nil || logger == nil {
		return nil, model.ErrGetActiveQueriesUC
	}
	return &DeleteAnswer{QueryService: queryService, logger: logger}, nil
}

func (uc *DeleteAnswer) DeleteAnswer(userID int, Query_name string) error {
	uc.logger.Info("DeleteAnswer", "userID", userID)
	req := &querypb.DeleteAnswerRequest{
		UserId:    int64(userID),
		QueryName: Query_name,
	}

	_, err := uc.QueryService.DeleteAnswer(context.Background(), req)
	if err != nil {
		return err
	}

	return err
}
