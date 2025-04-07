package usecase

import (
	"context"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type UserLogIn struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	hasher      repository.PasswordHasher
}

func NewUserLogInUseCase(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	hasher repository.PasswordHasher,
) *UserLogIn {
	return &UserLogIn{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		hasher:      hasher,
	}
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

	if !uc.hasher.Compare(user.Password, input.Password) {
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
	return uc.sessionRepo.StoreSession(session.SessionId, strconv.Itoa(session.UserId), session.Expires)
}

func (uc *UserLogIn) CreateCookies(ctx context.Context, session model.Session) (*model.Cookie, error) {
	cookie := &model.Cookie{
		Name:     "session_id",
		Value:    session.SessionId,
		HttpOnly: true,
		Secure:   false,
		Expires:  time.Now().Add(session.Expires),
		Path:     "/",
	}
	return cookie, nil
}

func (uc *UserLogIn) GetSession(sessionId string) (string, error) {
	return uc.sessionRepo.GetSession(sessionId)
}

type UserSignUp struct {
}

type UserCheckSession struct {
}

type UserLogOut struct {
}

type UserDeleteById struct {
}

type GetProfileById struct {
}

type GetProfilesForUser struct {
}
