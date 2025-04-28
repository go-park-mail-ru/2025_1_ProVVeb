package usecase

import (
	"context"
	"math/rand"
	"time"

	"github.com/icrowley/fake"
	"github.com/sirupsen/logrus"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type UserSignUp struct {
	userRepo  repository.UserRepository
	statRepo  repository.StaticRepository
	hasher    repository.PasswordHasher
	validator repository.UserParamsValidator
	logger    *logger.LogrusLogger
}

func NewUserSignUpUseCase(
	userRepo repository.UserRepository,
	statRepo repository.StaticRepository,
	hasher repository.PasswordHasher,
	validator repository.UserParamsValidator,
	logger *logger.LogrusLogger,
) (*UserSignUp, error) {
	if userRepo == nil || statRepo == nil || hasher == nil || validator == nil || logger == nil {
		return nil, model.ErrUserSignUpUC
	}
	return &UserSignUp{
		userRepo:  userRepo,
		statRepo:  statRepo,
		hasher:    hasher,
		validator: validator,
		logger:    logger,
	}, nil
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
	is := uc.userRepo.UserExists(ctx, login)
	uc.logger.WithFields(&logrus.Fields{"login": login, "is": is}).Info("UserExists")
	return is
}

func (uc *UserSignUp) SaveUserData(userId int, login, password string) (int, error) {
	uc.logger.WithFields(&logrus.Fields{"login": login, "password": password}).Info("SaveUserData")
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

	result, err := uc.userRepo.StoreUser(user)
	uc.logger.WithFields(&logrus.Fields{"result": result, "error": err, "user": user}).Info("SaveUserData")
	return result, err
}

func (uc *UserSignUp) SaveUserProfile(login string) (int, error) {
	uc.logger.WithFields(&logrus.Fields{"login": login}).Info("SaveUserProfile")
	var fname string
	var lname string

	ismale := (rand.Intn(2) == 0)

	if ismale {
		fname = fake.MaleFirstName()
		lname = fake.MaleLastName()
	} else {
		fname = fake.FemaleFirstName()
		lname = fake.FemaleLastName()
	}

	birthdate := time.Now().AddDate(-rand.Int()%27-18, -rand.Int()%12, -rand.Int()%30)
	height := rand.Int()%100 + 100
	description := fake.SentencesN(2)
	location := fake.City()
	interests := make([]string, 0, 5)
	for range 5 {
		interests = append(interests, fake.Word())
	}

	photos := make([]string, 0, 6)
	defaultFileName := "/" + fake.CharactersN(15) + ".png"
	photos = append(photos, defaultFileName)

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

	uc.logger.Info("Profile data generated")

	imgBytes, err := uc.statRepo.GenerateImage("image/png", ismale)
	if err != nil {
		uc.logger.Error("cannot generate image", err)
		return -1, err
	}

	err = uc.statRepo.UploadImage(imgBytes, defaultFileName, "image/png")
	if err != nil {
		uc.logger.Error("cannot upload image", err)
		return -1, err
	}

	profileId, err := uc.userRepo.StoreProfile(profile)
	if err != nil {
		uc.logger.Error("cannot store profile", err)
		return -1, err
	}

	err = uc.userRepo.StorePhotos(profileId, photos)
	if err != nil {
		uc.logger.Error("cannot store photos", err)
		return -1, err
	}

	err = uc.userRepo.StoreInterests(profileId, interests)
	if err != nil {
		uc.logger.Error("cannot store interests", err)
		return -1, err
	}
	uc.logger.WithFields(&logrus.Fields{"profileId": profileId}).Info("Profile saved")

	return profileId, nil
}
