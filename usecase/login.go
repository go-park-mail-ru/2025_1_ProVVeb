package usecase

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/durationpb"

	sessionpb "github.com/go-park-mail-ru/2025_1_ProVVeb/auth_micro/proto"
	userspb "github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/delivery"
)

type UserLogIn struct {
	hasher         repository.PasswordHasher
	token          repository.JwtTokenizer
	UsersService   userspb.UsersServiceClient
	SessionService sessionpb.SessionServiceClient
	logger         *logger.LogrusLogger
}

func NewUserLogInUseCase(
	hasher repository.PasswordHasher,
	token repository.JwtTokenizer,
	UsersService userspb.UsersServiceClient,
	SessionService sessionpb.SessionServiceClient,
	logger *logger.LogrusLogger,
) (*UserLogIn, error) {
	if hasher == nil {
		return nil, model.ErrUserLogInUC
	}
	return &UserLogIn{
		hasher: hasher,
		token:  token,

		UsersService:   UsersService,
		SessionService: SessionService,
		logger:         logger,
	}, nil
}

type LogInInput struct {
	Login    string
	Password string
}

func (uc *UserLogIn) CreateSession(ctx context.Context, input LogInInput) (model.Session, error) {
	login_req := &userspb.GetUserByLoginRequest{Login: input.Login}
	res, err := uc.UsersService.GetUserByLogin(context.Background(), login_req)
	if err != nil {
		uc.logger.Error("GetUserParams", "error", err)
		return model.Session{}, err
	}
	user := model.User{
		UserId:   int(res.User.UserId),
		Login:    res.User.Login,
		Password: res.User.Password,
		Email:    res.User.Email,
		Phone:    res.User.Phone,
		Status:   int(res.User.Status),
	}
	if !uc.hasher.Compare(user.Password, user.Login, input.Password) {
		uc.logger.Error("CreateSession", "error", model.ErrInvalidPassword)
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
		uc.logger.WithFields(&logrus.Fields{"session": session, "error": err}).Error("StoreSession")
		return err
	}

	uc.logger.WithFields(&logrus.Fields{"session": session, "error": err}).Error("StoreSession")
	return err
}

func (uc *UserLogIn) CreateJwtToken(s *repository.Session, tokenExpTime int64) (string, error) {
	result, err := uc.token.CreateJwtToken(s, tokenExpTime)
	uc.logger.WithFields(&logrus.Fields{"session": s, "tokenExpTime": tokenExpTime, "result": result, "error": err}).Info("CreateJwtToken")
	return result, err
}

func (uc *UserLogIn) GetSession(sessionId string) (string, error) {
	req := &sessionpb.SessionIdRequest{
		SessionId: sessionId,
	}

	sessionResp, err := uc.SessionService.GetSession(context.Background(), req)
	uc.logger.WithFields(&logrus.Fields{"sessionId": sessionId, "result": sessionResp, "error": err}).Info("GetSession")
	if err != nil {
		return "", err
	}

	return sessionResp.Data, nil
}
