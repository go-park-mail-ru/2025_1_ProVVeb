package usecase

import (
	"context"
	"math/rand"
	"strconv"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type UserLogIn struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	hasher      repository.PasswordHasher
	validator   repository.UserParamsValidator
}

func NewUserLogInUseCase(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	hasher repository.PasswordHasher,
	validator repository.UserParamsValidator,
) (*UserLogIn, error) {
	if userRepo == nil || sessionRepo == nil || hasher == nil || validator == nil {
		return nil, model.ErrUserLogInUC
	}
	return &UserLogIn{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		hasher:      hasher,
		validator:   validator,
	}, nil
}

type LogInInput struct {
	Login    string
	Password string
}

func RandStringRunes(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (uc *UserLogIn) CreateSession(ctx context.Context, input LogInInput) (model.Session, error) {
	user, err := uc.userRepo.GetUserByLogin(ctx, input.Login)
	if err != nil {
		return model.Session{}, err
	}

	if !uc.hasher.Compare(user.Password, user.Login, input.Password) {
		return model.Session{}, model.ErrInvalidPassword
	}

	session_id := RandStringRunes(model.SessionIdLength)
	expires := model.SessionDuration

	session := model.Session{
		SessionId: session_id,
		UserId:    user.UserId,
		Expires:   expires,
	}

	return session, nil
}

func (uc *UserLogIn) StoreSession(ctx context.Context, session model.Session) error {
	err := uc.sessionRepo.StoreSession(session.SessionId, strconv.Itoa(session.UserId), session.Expires)
	if err != nil {
		return err
	}

	err = uc.userRepo.StoreSession(session.UserId, session.SessionId)
	return err
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
