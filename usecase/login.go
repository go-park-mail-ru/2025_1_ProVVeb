package usecase

import (
	"context"
	"strconv"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type UserLogIn struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	hasher      repository.PasswordHasher
	token       repository.JwtToken
	validator   repository.UserParamsValidator
}

func NewUserLogInUseCase(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	hasher repository.PasswordHasher,
	token repository.JwtToken,
	validator repository.UserParamsValidator,
) (*UserLogIn, error) {
	if userRepo == nil || sessionRepo == nil || hasher == nil || validator == nil {
		return nil, model.ErrUserLogInUC
	}
	return &UserLogIn{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		hasher:      hasher,
		token:       token,
		validator:   validator,
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

	session := uc.sessionRepo.CreateSession(user.UserId)

	return session, nil
}

func (uc *UserLogIn) CheckAttempts(ctx context.Context, userIP string) error {
	return uc.sessionRepo.CheckAttempts(userIP)
}

func (uc *UserLogIn) IncreaseAttempts(ctx context.Context, userIP string) error {
	return uc.sessionRepo.IncreaseAttempts(userIP)
}

func (uc *UserLogIn) DeleteAttempts(ctx context.Context, userIP string) error {
	return uc.sessionRepo.DeleteAttempts(userIP)
}

func (uc *UserLogIn) StoreSession(ctx context.Context, session model.Session) error {
	err := uc.sessionRepo.StoreSession(session.SessionId, strconv.Itoa(session.UserId), session.Expires)
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
	return uc.sessionRepo.GetSession(sessionId)
}

func (uc *UserLogIn) ValidateLogin(login string) bool {
	return uc.validator.ValidateLogin(login) == nil
}

func (uc *UserLogIn) ValidatePassword(password string) bool {
	return uc.validator.ValidatePassword(password) == nil
}
