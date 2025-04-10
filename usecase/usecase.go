package usecase

import (
	"context"
	"math/rand"
	"strconv"
	"time"

	"github.com/icrowley/fake"

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
) *UserLogIn {
	return &UserLogIn{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		hasher:      hasher,
		validator:   validator,
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

func (uc *UserLogIn) ValidateLogin(login string) bool {
	return uc.validator.ValidateLogin(login) == nil
}

func (uc *UserLogIn) ValidatePassword(password string) bool {
	return uc.validator.ValidatePassword(password) == nil
}

type UserSignUp struct {
	userRepo  repository.UserRepository
	hasher    repository.PasswordHasher
	validator repository.UserParamsValidator
}

func NewUserSignUpUseCase(
	userRepo repository.UserRepository,
	hasher repository.PasswordHasher,
	validator repository.UserParamsValidator,
) *UserSignUp {
	return &UserSignUp{
		userRepo:  userRepo,
		hasher:    hasher,
		validator: validator,
	}
}

type UserSignUpInput struct {
	Login    string
	Password string
}

func (uc *UserSignUp) ValidateLogin(login string) error {
	return uc.validator.ValidateLogin(login)
}

func (uc *UserSignUp) ValidatePassword(password string) error {
	return uc.validator.ValidatePassword(password)
}

func (uc *UserSignUp) UserExists(ctx context.Context, login string) bool {
	return uc.userRepo.UserExists(ctx, login)
}

func (uc *UserSignUp) SaveUserData(userId int, login, password string) (int, error) {
	email := fake.EmailAddress()
	phone := fake.Phone()
	status := 0
	user := model.User{
		Login:    login,
		Password: uc.hasher.Hash(password),
		Email:    email,
		Phone:    phone,
		Status:   status,
		UserId:   userId,
	}

	return uc.userRepo.StoreUser(user)
}

func (uc *UserSignUp) SaveUserProfile(login string) (int, error) {
	fname := fake.FirstName()
	lname := fake.LastName()
	ismale := true
	birthdate, _ := time.Parse("2006-01-02", "1990-01-01")
	height := rand.Int()%100 + 100
	description := fake.SentencesN(5)

	profile := model.Profile{
		FirstName:   fname,
		LastName:    lname,
		IsMale:      ismale,
		Birthday:    birthdate,
		Height:      height,
		Description: description,
	}

	return uc.userRepo.StoreProfile(profile)
}

type UserCheckSession struct {
	sessionRepo repository.SessionRepository
}

func NewUserCheckSessionUseCase(sessionRepo repository.SessionRepository) *UserCheckSession {
	return &UserCheckSession{sessionRepo: sessionRepo}
}

func (uc *UserCheckSession) CheckSession(sessionId string) (int, error) {
	userIdStr, err := uc.sessionRepo.GetSession(sessionId)
	if err != nil {
		return -1, err
	}

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return -1, model.ErrInvalidSessionId
	}
	return userId, nil
}

type UserLogOut struct {
}

type UserDeleteById struct {
}

type GetProfileById struct {
}

type GetProfilesForUser struct {
}
