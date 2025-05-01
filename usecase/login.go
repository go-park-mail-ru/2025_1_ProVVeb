package usecase

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"google.golang.org/protobuf/types/known/durationpb"

	sessionpb "github.com/go-park-mail-ru/2025_1_ProVVeb/auth_micro/proto"
)

type UserLogIn struct {
	userRepo       repository.UserRepository
	sessionRepo    repository.SessionRepository
	hasher         repository.PasswordHasher
	token          repository.JwtTokenizer
	validator      repository.UserParamsValidator
	SessionService sessionpb.SessionServiceClient
}

func NewUserLogInUseCase(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	hasher repository.PasswordHasher,
	token repository.JwtTokenizer,
	validator repository.UserParamsValidator,
	SessionService sessionpb.SessionServiceClient,
) (*UserLogIn, error) {
	if userRepo == nil || sessionRepo == nil || hasher == nil || validator == nil {
		return nil, model.ErrUserLogInUC
	}
	return &UserLogIn{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		hasher:         hasher,
		token:          token,
		validator:      validator,
		SessionService: SessionService,
	}, nil
}

type LogInInput struct {
	Login    string
	Password string
}

func (uc *UserLogIn) CreateSession(ctx context.Context, input LogInInput) (model.Session, error) {
	user, err := uc.userRepo.GetUserByLogin(ctx, input.Login)
	if err != nil {
		return model.Session{}, err
	}

	if !uc.hasher.Compare(user.Password, user.Login, input.Password) {
		return model.Session{}, model.ErrInvalidPassword
	}

	req := &sessionpb.CreateSessionRequest{
		UserId: int32(user.UserId),
	}

	sessionResp, err := uc.SessionService.CreateSession(ctx, req)
	if err != nil {
		return model.Session{}, err
	}

	session := model.Session{
		SessionId: sessionResp.SessionId,
		UserId:    int(user.UserId),
	}

	return session, err
}

func (uc *UserLogIn) CheckAttempts(ctx context.Context, userIP string) (string, error) {
	ipRequest := &sessionpb.IPRequest{
		Ip: userIP,
	}

	sessionResp, err := uc.SessionService.CheckAttempts(ctx, ipRequest)
	fmt.Println(sessionResp, err)
	if err != nil {
		return "", err
	}

	if sessionResp.ErrorMessage != "" {
		return "", fmt.Errorf(sessionResp.ErrorMessage)
	}

	return sessionResp.BlockUntil, nil
}

func (uc *UserLogIn) IncreaseAttempts(ctx context.Context, userIP string) (string, error) {
	ipRequest := &sessionpb.IPRequest{
		Ip: userIP,
	}

	sessionResp, err := uc.SessionService.IncreaseAttempts(ctx, ipRequest)
	if err != nil {
		return "", err
	}

	return sessionResp.String(), nil
}

func (uc *UserLogIn) DeleteAttempts(ctx context.Context, userIP string) error {
	ipRequest := &sessionpb.IPRequest{
		Ip: userIP,
	}

	_, err := uc.SessionService.DeleteAttempts(ctx, ipRequest)
	if err != nil {
		return err
	}

	return nil
}

func (uc *UserLogIn) StoreSession(ctx context.Context, session model.Session) error {
	userIdStr := strconv.Itoa(session.UserId)

	ttl := durationpb.New(session.Expires)

	req := &sessionpb.StoreSessionRequest{
		SessionId: session.SessionId,
		Data:      userIdStr,
		Ttl:       ttl,
	}

	_, err := uc.SessionService.StoreSession(ctx, req)
	if err != nil {
		return err
	}

	err = uc.userRepo.StoreSession(session.UserId, session.SessionId)
	return err
}

func (uc *UserLogIn) CreateJwtToken(s *repository.Session, tokenExpTime int64) (string, error) {
	return uc.token.CreateJwtToken(s, tokenExpTime)
}

func (uc *UserLogIn) GetSession(sessionId string) (string, error) {
	req := &sessionpb.SessionIdRequest{
		SessionId: sessionId,
	}

	sessionResp, err := uc.SessionService.GetSession(context.Background(), req)
	if err != nil {
		return "", err
	}

	return sessionResp.Data, nil
}

func (uc *UserLogIn) ValidateLogin(login string) bool {
	return uc.validator.ValidateLogin(login) == nil
}

func (uc *UserLogIn) ValidatePassword(password string) bool {
	return uc.validator.ValidatePassword(password) == nil
}
