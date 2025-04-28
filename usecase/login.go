package usecase

import (
	"context"
	"strconv"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/sirupsen/logrus"
)

type UserLogIn struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	hasher      repository.PasswordHasher
	token       repository.JwtTokenizer
	validator   repository.UserParamsValidator
	logger      *logger.LogrusLogger
}

func NewUserLogInUseCase(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	hasher repository.PasswordHasher,
	token repository.JwtTokenizer,
	validator repository.UserParamsValidator,
	logger *logger.LogrusLogger,
) (*UserLogIn, error) {
	if userRepo == nil || sessionRepo == nil || hasher == nil || validator == nil || logger == nil {
		return nil, model.ErrUserLogInUC
	}
	return &UserLogIn{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		hasher:      hasher,
		token:       token,
		validator:   validator,
		logger:      logger,
	}, nil
}

type LogInInput struct {
	Login    string
	Password string
}

func (uc *UserLogIn) CreateSession(ctx context.Context, input LogInInput) (model.Session, error) {
	uc.logger.Info("CreateSession", "input", input)
	user, err := uc.userRepo.GetUserByLogin(ctx, input.Login)
	if err != nil {
		uc.logger.Error("CreateSession", "error", err)
		return model.Session{}, err
	}

	if !uc.hasher.Compare(user.Password, user.Login, input.Password) {
		uc.logger.Error("CreateSession", "error", model.ErrInvalidPassword)
		return model.Session{}, model.ErrInvalidPassword
	}

	uc.logger.Info("CreateSession", "userId", user.UserId)
	session := uc.sessionRepo.CreateSession(user.UserId)

	uc.logger.Info("CreateSession", "session", session)
	return session, nil
}

func (uc *UserLogIn) CheckAttempts(ctx context.Context, userIP string) error {
	err := uc.sessionRepo.CheckAttempts(userIP)
	uc.logger.WithFields(&logrus.Fields{"userIP": userIP, "error": err})
	return err
}

func (uc *UserLogIn) IncreaseAttempts(ctx context.Context, userIP string) error {
	err := uc.sessionRepo.IncreaseAttempts(userIP)
	uc.logger.WithFields(&logrus.Fields{"userIP": userIP, "error": err})
	return err
}

func (uc *UserLogIn) DeleteAttempts(ctx context.Context, userIP string) error {
	err := uc.sessionRepo.DeleteAttempts(userIP)
	uc.logger.WithFields(&logrus.Fields{"userIP": userIP, "error": err})
	return err
}

func (uc *UserLogIn) StoreSession(ctx context.Context, session model.Session) error {
	uc.logger.WithFields(&logrus.Fields{"session": session}).Info("StoreSession")
	err := uc.sessionRepo.StoreSession(session.SessionId, strconv.Itoa(session.UserId), session.Expires)
	if err != nil {
		uc.logger.WithFields(&logrus.Fields{"session": session, "error": err}).Error("StoreSession")
		return err
	}

	err = uc.userRepo.StoreSession(session.UserId, session.SessionId)
	uc.logger.WithFields(&logrus.Fields{"session": session, "error": err}).Error("StoreSession")
	return err
}

func (uc *UserLogIn) CreateJwtToken(s *repository.Session, tokenExpTime int64) (string, error) {
	result, err := uc.token.CreateJwtToken(s, tokenExpTime)
	uc.logger.WithFields(&logrus.Fields{"session": s, "tokenExpTime": tokenExpTime, "result": result, "error": err}).Info("CreateJwtToken")
	return result, err
}

func (uc *UserLogIn) GetSession(sessionId string) (string, error) {
	result, err := uc.sessionRepo.GetSession(sessionId)
	uc.logger.WithFields(&logrus.Fields{"sessionId": sessionId, "result": result, "error": err}).Info("GetSession")
	return result, err
}

func (uc *UserLogIn) ValidateLogin(login string) bool {
	return uc.validator.ValidateLogin(login) == nil
}

func (uc *UserLogIn) ValidatePassword(password string) bool {
	return uc.validator.ValidatePassword(password) == nil
}
