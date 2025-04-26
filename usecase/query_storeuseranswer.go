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

type userAnswer struct {
	Name   string `json:"name"`
	Score  int32  `json:"score"`
	Answer string `json:"answer"`
}

func (s *StoreUserAnswer) StoreUserAnswer(userID int32, answer userAnswer) error {
	req := &querypb.SendRespRequest{
		UserId: userID,
		Name:   answer.Name,
		Score:  answer.Score,
		Answer: answer.Answer,
	}

	_, err := s.QueryService.SendResp(context.Background(), req)

	return err
}
