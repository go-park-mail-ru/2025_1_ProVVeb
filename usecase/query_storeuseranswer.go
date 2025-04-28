package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	querypb "github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/proto"
	"github.com/sirupsen/logrus"
)

type StoreUserAnswer struct {
	QueryService querypb.QueryServiceClient
	logger       *logger.LogrusLogger
}

func NewStoreUserAnswer(
	queryService querypb.QueryServiceClient,
	logger *logger.LogrusLogger,
) (*StoreUserAnswer, error) {
	if queryService == nil || logger == nil {
		return nil, model.ErrGetActiveQueriesUC
	}
	return &StoreUserAnswer{QueryService: queryService, logger: logger}, nil
}

func (s *StoreUserAnswer) StoreUserAnswer(userID int32, name string, score int32, answer string) error {
	s.logger.WithFields(&logrus.Fields{"userID": userID, "name": name, "score": score, "answer": answer})
	req := &querypb.SendRespRequest{
		UserId: userID,
		Name:   name,
		Score:  score,
		Answer: answer,
	}

	_, err := s.QueryService.SendResp(context.Background(), req)
	s.logger.WithFields(&logrus.Fields{"error": err}).Error("StoreUserAnswer")

	return err
}
