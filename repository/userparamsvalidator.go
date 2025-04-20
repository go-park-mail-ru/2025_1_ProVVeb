package repository

import (
	"fmt"
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

func (vr *UParamsValidator) ValidateLogin(login string) error {
	if (len(login) < model.MinLoginLength) || (len(login) > model.MaxLoginLength) {
		return fmt.Errorf("incorrect size of login")
	}

	re := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]*$`)
	if !re.MatchString(login) {
		return fmt.Errorf("incorrect format of login")
	}
	return nil
}

func (vr *UParamsValidator) ValidatePassword(password string) error {
	if (len(password) < model.MinPasswordLength) || (len(password) > model.MaxPasswordLength) {
		return fmt.Errorf("incorrect size of password")
	}
	// ideas for future
	// password must contain at least one digit
	// password must contain only letters and digits
	// password must contain at least one special character
	// password must not contain invalid characters

	return nil
}
