package usecase

import (
	"context"

	querypb "github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/proto"
)

type StoreUserAnswer struct {
	QueryService querypb.QueryServiceClient
}

func NewStoreUserAnswer(queryService querypb.QueryServiceClient) *StoreUserAnswer {
	return &StoreUserAnswer{QueryService: queryService}
}

func (s *StoreUserAnswer) StoreUserAnswer(userID int32, name string, score int32, answer string) error {
	req := &querypb.SendRespRequest{
		UserId: userID,
		Name:   name,
		Score:  score,
		Answer: answer,
	}

	_, err := s.QueryService.SendResp(context.Background(), req)

	return err
}
