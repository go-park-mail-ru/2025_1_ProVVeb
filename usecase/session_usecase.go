package usecase

import (
	"context"
	"strconv"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
)

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
