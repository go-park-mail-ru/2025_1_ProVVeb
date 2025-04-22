package usecase

import (
	"context"
	"math/rand"
	"time"

	"github.com/icrowley/fake"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
)

type UserSignUp struct {
	userRepo  repository.UserRepository
	statRepo  repository.StaticRepository
	hasher    repository.PasswordHasher
	validator repository.UserParamsValidator
}

func NewUserSignUpUseCase(
	userRepo repository.UserRepository,
	statRepo repository.StaticRepository,
	hasher repository.PasswordHasher,
	validator repository.UserParamsValidator,
) (*UserSignUp, error) {
	if userRepo == nil || statRepo == nil || hasher == nil || validator == nil {
		return nil, model.ErrUserSignUpUC
	}
	return &UserSignUp{
		userRepo:  userRepo,
		statRepo:  statRepo,
		hasher:    hasher,
		validator: validator,
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
	return uc.userRepo.UserExists(ctx, login)
}

func (uc *UserSignUp) SaveUserData(userId int, sent_user model.User) (int, error) {
	var email string
	if sent_user.Email == "" {
		email = fake.EmailAddress()
	} else {
		email = sent_user.Email
	}

	var phone string
	if sent_user.Phone == "" {
		phone = fake.Phone()
	} else {
		phone = sent_user.Phone
	}
	status := 0
	user := model.User{
		Login:    sent_user.Login,
		Password: uc.hasher.Hash(sent_user.Login + "_" + sent_user.Password),
		Email:    email,
		Phone:    phone,
		Status:   status,
		UserId:   userId,
	}

	return uc.userRepo.StoreUser(user)
}

func (uc *UserSignUp) SaveUserProfile(sent_profile model.Profile) (int, error) {
	var fname, lname string
	if sent_profile.FirstName != "" {
		fname = sent_profile.FirstName
	} else {
		if sent_profile.IsMale {
			fname = fake.MaleFirstName()
		} else {
			fname = fake.FemaleFirstName()
		}
	}

	if sent_profile.LastName != "" {
		lname = sent_profile.LastName
	} else {
		if sent_profile.IsMale {
			lname = fake.MaleLastName()
		} else {
			lname = fake.FemaleLastName()
		}
	}

	var birthdate time.Time
	if sent_profile.Birthday.IsZero() {
		birthdate = time.Now().AddDate(-(rand.Intn(27) + 18), -rand.Intn(12), -rand.Intn(30))
	} else {
		birthdate = sent_profile.Birthday
	}

	height := sent_profile.Height
	if height == 0 {
		height = rand.Intn(100) + 100
	}

	description := sent_profile.Description
	if description == "" {
		description = fake.SentencesN(2)
	}

	location := sent_profile.Location
	if location == "" {
		location = fake.City()
	}

	interests := sent_profile.Interests
	if len(interests) == 0 {
		for i := 0; i < 5; i++ {
			interests = append(interests, fake.Word())
		}
	}

	photos := make([]string, 0, 6)
	defaultFileName := "/" + fake.CharactersN(15) + ".png"
	photos = append(photos, defaultFileName)

	profile := model.Profile{
		FirstName:   fname,
		LastName:    lname,
		IsMale:      sent_profile.IsMale,
		Birthday:    birthdate,
		Height:      height,
		Description: description,
		Location:    location,
		Interests:   interests,
		Photos:      photos,
		Preferences: sent_profile.Preferences,
		LikedBy:     sent_profile.LikedBy,
	}

	// fmt.Println(fmt.Errorf("profile: %+v", profile))

	imgBytes, err := uc.statRepo.GenerateImage("image/png", sent_profile.IsMale)
	if err != nil {
		return -1, err
	}

	err = uc.statRepo.UploadImage(imgBytes, defaultFileName, "image/png")
	if err != nil {
		return -1, err
	}

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
