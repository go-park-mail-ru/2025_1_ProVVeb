package auth

import (
	"context"
	"fmt"
	"time"

	sessionpb "github.com/go-park-mail-ru/2025_1_ProVVeb/auth_micro/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	if err := s.Repo.StoreSession(session.UserId, session.SessionId, "session_data", time.Duration(session.Expires)); err != nil {
		return nil, fmt.Errorf("error storing session: %v", err)
	}
	expiresDuration := durationpb.New(12 * time.Hour)

	sessionResponse := &sessionpb.SessionResponse{
		SessionId: session.SessionId,
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

func (s *SessionServiceServerImpl) StoreSession(ctx context.Context, req *sessionpb.StoreSessionRequest) (*emptypb.Empty, error) {
	ttl := req.Ttl.AsDuration()

	err := s.Repo.StoreSession(0, req.Data, req.SessionId, ttl)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to store session: %v", err)
	}

	return &emptypb.Empty{}, nil
}
