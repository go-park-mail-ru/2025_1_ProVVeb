package usecase

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/icrowley/fake"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

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
		Password: uc.hasher.Hash(login + "_" + password),
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
	location := fake.City()
	interests := make([]string, 0, 20)
	for range 20 {
		interests = append(interests, fake.Word())
	}
	photos := make([]string, 0, 6)
	photos = append(photos, "default.png")

	profile := model.Profile{
		FirstName:   fname,
		LastName:    lname,
		IsMale:      ismale,
		Birthday:    birthdate,
		Height:      height,
		Description: description,
		Interests:   interests,
		Location:    location,
		Photos:      photos,
	}

	fmt.Println(fmt.Errorf("profile: %+v", profile))

	profileId, err := uc.userRepo.StoreProfile(profile)
	if err != nil {
		return -1, err
	}

	err = uc.userRepo.StorePhotos(profileId, photos)
	if err != nil {
		return -1, err
	}

	err = uc.userRepo.StoreInterests(profileId, interests)
	if err != nil {
		return -1, err
	}

	return profileId, nil
}
