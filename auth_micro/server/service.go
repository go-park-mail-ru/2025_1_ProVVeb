package auth

import (
	"context"
	"fmt"
	"time"

	sessionpb "github.com/go-park-mail-ru/2025_1_ProVVeb/auth_micro/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type SessionServiceServerImpl struct {
	sessionpb.UnimplementedSessionServiceServer
	Repo *SessionRepo
}

func NewSessionService(repo *SessionRepo) *SessionServiceServerImpl {
	return &SessionServiceServerImpl{Repo: repo}
}

func (s *SessionServiceServerImpl) CreateSession(ctx context.Context, req *sessionpb.CreateSessionRequest) (*sessionpb.SessionResponse, error) {
	session := s.Repo.CreateSession(int(req.GetUserId()))
	if err := s.Repo.StoreSession(session.SessionId, "session_data", time.Duration(session.Expires)*time.Second); err != nil {
		return nil, fmt.Errorf("error storing session: %v", err)
	}
	expiresDuration := durationpb.New(3600 * time.Second)

	sessionResponse := &sessionpb.SessionResponse{
		SessionId: "some_session_id",
		UserId:    req.GetUserId(),
		Expires:   expiresDuration,
	}
	return sessionResponse, nil
}

func (s *SessionServiceServerImpl) GetSession(ctx context.Context, req *sessionpb.SessionIdRequest) (*sessionpb.SessionDataResponse, error) {
	data, err := s.Repo.GetSession(req.GetSessionId())
	if err != nil {
		return nil, fmt.Errorf("error getting session: %v", err)
	}

	return &sessionpb.SessionDataResponse{
		Data: data,
	}, nil
}

func (s *SessionServiceServerImpl) DeleteSession(ctx context.Context, req *sessionpb.SessionIdRequest) (*emptypb.Empty, error) {
	if err := s.Repo.DeleteSession(req.GetSessionId()); err != nil {
		return nil, fmt.Errorf("error deleting session: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *SessionServiceServerImpl) CheckAttempts(ctx context.Context, req *sessionpb.IPRequest) (*sessionpb.CheckAttemptsResponse, error) {
	blockTime, err := s.Repo.CheckAttempts(req.GetIp())
	if err != nil {
		return nil, fmt.Errorf("error checking attempts: %v", err)
	}

	return &sessionpb.CheckAttemptsResponse{
		BlockUntil: blockTime,
	}, nil
}

func (s *SessionServiceServerImpl) IncreaseAttempts(ctx context.Context, req *sessionpb.IPRequest) (*emptypb.Empty, error) {
	if err := s.Repo.IncreaseAttempts(req.GetIp()); err != nil {
		return nil, fmt.Errorf("error increasing attempts: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *SessionServiceServerImpl) DeleteAttempts(ctx context.Context, req *sessionpb.IPRequest) (*emptypb.Empty, error) {
	if err := s.Repo.DeleteAttempts(req.GetIp()); err != nil {
		return nil, fmt.Errorf("error deleting attempts: %v", err)
	}

	return &emptypb.Empty{}, nil
}
