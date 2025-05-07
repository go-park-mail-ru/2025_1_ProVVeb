package repository

import (
	"regexp"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
)

type UserParamsValidator interface {
	ValidateLogin(login string) error
	ValidatePassword(password string) error
}

type UParamsValidator struct{}

func NewUParamsValidator() (*UParamsValidator, error) {
	return &UParamsValidator{}, nil
}

func (uv *UParamsValidator) ValidateLogin(login string) error {
	if (len(login) < model.MinLoginLength) || (len(login) > model.MaxLoginLength) {
		return model.ErrInvalidLoginSize
	}

	re := regexp.MustCompile(model.ReStartsWithLetter)
	if !re.MatchString(login) {
		return model.ErrInvalidLogin
	}

	re = regexp.MustCompile(model.ReContainsLettersDigitsSymbols)
	if !re.MatchString(login) {
		return model.ErrInvalidLogin
	}

	return nil
}

func (uv *UParamsValidator) ValidatePassword(password string) error {
	if (len(password) < model.MinPasswordLength) || (len(password) > model.MaxPasswordLength) {
		return model.ErrInvalidPasswordSize
	}

	return nil
}
