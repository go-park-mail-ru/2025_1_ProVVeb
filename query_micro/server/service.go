package query

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/config"
	querypb "github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type QueryServiceServerImpl struct {
	querypb.UnimplementedQueryServiceServer
	Repo *QueryRepo
}

func NewSessionService(repo *QueryRepo) *QueryServiceServerImpl {
	return &QueryServiceServerImpl{Repo: repo}
}

func (s *QueryServiceServerImpl) GetActive(ctx context.Context, req *querypb.GetUserRequest) (*querypb.ActiveQueryList, error) {
	queries, err := s.Repo.GetActive(int(req.GetUserId()))
	if err != nil {
		return nil, fmt.Errorf("error getting active queries: %v", err)
	}

	var activeQueries []*querypb.ActiveQuery
	for _, query := range queries {
		activeQueries = append(activeQueries, &querypb.ActiveQuery{
			Name:        query.Name,
			Description: query.Description,
			MinScore:    int32(query.MinScore),
			MaxScore:    int32(query.MaxScore),
		})
	}

	return &querypb.ActiveQueryList{Items: activeQueries}, nil
}

func (s *QueryServiceServerImpl) SendResp(ctx context.Context, req *querypb.SendRespRequest) (*emptypb.Empty, error) {
	answer := config.Answer{
		QueryName: req.GetName(),
		UserId:    int(req.GetUserId()),
		Score:     int(req.GetScore()),
		Answer:    req.GetAnswer(),
	}

	err := s.Repo.SendResp(answer)
	if err != nil {
		return nil, fmt.Errorf("error sending response: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *QueryServiceServerImpl) GetForUser(ctx context.Context, req *querypb.GetUserRequest) (*querypb.QueryResponseList, error) {
	answers, err := s.Repo.GetForUser(int(req.GetUserId()))
	if err != nil {
		return nil, fmt.Errorf("error getting answers for user: %v", err)
	}

	var queryResponses []*querypb.QueryResponse
	for _, answer := range answers {
		queryResponses = append(queryResponses, &querypb.QueryResponse{
			Name:        answer.Name,
			Description: answer.Description,
			MinScore:    int32(answer.MinScore),
			MaxScore:    int32(answer.MaxScore),
			Answer:      answer.Answer,
			Score:       int32(answer.Score),
		})
	}

	return &querypb.QueryResponseList{Items: queryResponses}, nil
}

func (s *QueryServiceServerImpl) GetForQuery(ctx context.Context, req *emptypb.Empty) (*querypb.ForQueryResponseList, error) {
	usersForQueries, err := s.Repo.GetUsersForQueries()
	if err != nil {
		return nil, fmt.Errorf("error getting users for queries: %v", err)
	}

	var usersForQueryList []*querypb.ForQueryResponse
	for _, userForQuery := range usersForQueries {
		usersForQueryList = append(usersForQueryList, &querypb.ForQueryResponse{
			Name:        userForQuery.Name,
			Description: userForQuery.Description,
			MinScore:    int32(userForQuery.MinScore),
			MaxScore:    int32(userForQuery.MaxScore),
			Login:       userForQuery.Login,
			Answer:      userForQuery.Answer,
			Score:       int32(userForQuery.Score),
		})
	}

	return &querypb.ForQueryResponseList{Items: usersForQueryList}, nil
}

func (s *QueryServiceServerImpl) FindQuery(ctx context.Context, req *querypb.FindQueryRequest) (*querypb.FindQueryResponseList, error) {
	usersForQueries, err := s.Repo.FindQuery(req.Name, int(req.QueryId))
	if err != nil {
		return nil, fmt.Errorf("error getting users for queries: %v", err)
	}

	var usersForQueryList []*querypb.FindQueryResponse
	for _, userForQuery := range usersForQueries {
		usersForQueryList = append(usersForQueryList, &querypb.FindQueryResponse{
			Name:        userForQuery.Name,
			Description: userForQuery.Description,
			MinScore:    int32(userForQuery.MinScore),
			MaxScore:    int32(userForQuery.MaxScore),
			Login:       userForQuery.Login,
			Answer:      userForQuery.Answer,
			Score:       int32(userForQuery.Score),
			UserId:      int32(userForQuery.UserId),
		})
	}

	return &querypb.FindQueryResponseList{Items: usersForQueryList}, nil
}

func (s *QueryServiceServerImpl) DeleteAnswer(ctx context.Context, req *querypb.DeleteAnswerRequest) (*emptypb.Empty, error) {
	err := s.Repo.DeleteAnswer(req.QueryName, int(req.UserId))
	if err != nil {
		return nil, fmt.Errorf("error sending response: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *QueryServiceServerImpl) GetQueryStats(ctx context.Context, req *querypb.QueryStatsRequest) (*querypb.QueryStatsResponse, error) {
	stats, err := s.Repo.GetStatistics(req.QueryName)
	if err != nil {
		return nil, fmt.Errorf("error getting answers for user: %v", err)
	}

	return &querypb.QueryStatsResponse{
		TotalAnswers: stats.TotalAnswers,
		AverageScore: stats.AverageScore,
		MinScore:     int32(stats.MinScore),
		MaxScore:     int32(stats.MaxScore),
	}, nil
}
