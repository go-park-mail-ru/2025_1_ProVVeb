package usecase

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	profilespb "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	userspb "github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/delivery"
)

type UserSignUp struct {
	UsersService    userspb.UsersServiceClient
	ProfilesService profilespb.ProfilesServiceClient
	logger          *logger.LogrusLogger
}

func NewUserSignUpUseCase(
	UsersService userspb.UsersServiceClient,
	ProfilesService profilespb.ProfilesServiceClient,
	logger *logger.LogrusLogger,
) (*UserSignUp, error) {
	if UsersService == nil || ProfilesService == nil || logger == nil {
		return nil, model.ErrUserSignUpUC
	}
	return &UserSignUp{
		UsersService:    UsersService,
		ProfilesService: ProfilesService,
		logger:          logger,
	}, nil
}

type UserSignUpInput struct {
	Login    string
	Password string
}

func (uc *UserSignUp) ValidateLogin(login string) error {
	uc.logger.Info("ValidateLogin", "login", login)
	req := &userspb.ValidateLoginRequest{
		Login: login,
	}
	_, err := uc.UsersService.ValidateLogin(context.Background(), req)
	fmt.Println(login, err)
	uc.logger.Info("error", err)
	return err
}

func (uc *UserSignUp) ValidatePassword(password string) error {
	uc.logger.Info("ValidatePassword")
	req := &userspb.ValidatePasswordRequest{
		Password: password,
	}
	_, err := uc.UsersService.ValidatePassword(context.Background(), req)
	uc.logger.Info("error", err)
	return err
}

func (uc *UserSignUp) UserExists(login string) bool {
	uc.logger.Info("UserExists", "login", login)

	req := &userspb.UserExistsRequest{
		Login: login,
	}
	res, err := uc.UsersService.UserExists(context.Background(), req)
	if err != nil {
		uc.logger.Error("UserExists", "error", err)
		return false
	}

	is := res.Exists
	uc.logger.WithFields(&logrus.Fields{"login": login, "is": is}).Info("UserExists")
	return is
}

func (uc *UserSignUp) SaveUserData(userId int, sentUser model.User) (int, error) {
	uc.logger.WithFields(&logrus.Fields{
		"login":  sentUser.Login,
		"userId": sentUser.UserId,
	}).Info("SaveUserData")
	req := &userspb.SaveUserDataRequest{
		UserId: int32(userId),
		User: &userspb.User{
			UserId:   int32(sentUser.UserId),
			Login:    sentUser.Login,
			Password: sentUser.Password,
			Email:    sentUser.Email,
			Phone:    sentUser.Phone,
			Status:   int32(sentUser.Status),
		},
	}

	res, err := uc.UsersService.SaveUserData(context.Background(), req)
	if err != nil {
		uc.logger.WithFields(&logrus.Fields{"err": err}).Error("SaveUserData")
		return -1, err
	}
	result := int(res.UserId)
	return result, err
}

func (uc *UserSignUp) SaveUserProfile(sentProfile model.Profile) (int, error) {
	uc.logger.WithFields(&logrus.Fields{"login": sentProfile.FirstName}).Info("SaveUserProfile")

	likedBy := []int32{}
	for _, like := range sentProfile.LikedBy {
		likedBy = append(likedBy, int32(like))
	}

	prefs := []*profilespb.Preference{}
	for _, pref := range sentProfile.Preferences {
		prefs = append(prefs, &profilespb.Preference{
			Description: pref.Description,
			Value:       pref.Value,
		})
	}

	req := &profilespb.StoreProfileRequest{
		Profile: &profilespb.Profile{
			ProfileId:   int32(sentProfile.ProfileId),
			FirstName:   sentProfile.FirstName,
			LastName:    sentProfile.LastName,
			IsMale:      sentProfile.IsMale,
			Height:      int32(sentProfile.Height),
			Birthday:    timestamppb.New(sentProfile.Birthday),
			Description: sentProfile.Description,
			Location:    sentProfile.Location,
			Interests:   sentProfile.Interests,
			Photos:      sentProfile.Photos,
			LikedBy:     likedBy,
			Preferences: prefs,
		},
	}

	res, err := uc.ProfilesService.StoreProfile(context.Background(), req)
	uc.logger.WithFields(&logrus.Fields{"err": err, "profileId": int(res.GetProfileId())})
	profileId := int(res.ProfileId)

	return profileId, nil
}
