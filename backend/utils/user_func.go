package utils

import (
	"crypto/sha256"
	"fmt"
	"regexp"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/config"
)

type User struct {
	User config.User
}

func ValidateLogin(login string) error {
	if (len(login) < 7) || (len(login) > 15) {
		return fmt.Errorf("incorrect size of login")
	}

	re := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]$`)
	if !re.MatchString(login) {
		return fmt.Errorf("incorrect format of login")
	}
	return nil
}

func ValidatePassword(password string) error {
	if (len(password) < 8) || (len(password) > 50) {
		return fmt.Errorf("incorrect size of password")
	}

	reDigit := regexp.MustCompile(`[0-9]`)
	if !reDigit.MatchString(password) {
		return fmt.Errorf("password must contain at least one digit")
	}

	reSpecial := regexp.MustCompile(`[!@#$%^&*]`)
	if !reSpecial.MatchString(password) {
		return fmt.Errorf("password must contain at least one special character")
	}

	reValidChars := regexp.MustCompile(`^[A-Za-z\d!@#$%^&*]{8,50}$`)
	if !reValidChars.MatchString(password) {
		return fmt.Errorf("password contains invalid characters")
	}
	return nil
}

func (u User) PrintUser() string {
	return fmt.Sprintf("Current user ID: %d, Login: %s", u.User.Id, u.User.Login)
}

func EncryptPasswordSHA256(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func CreateUser(id int, login string, password string) (User, error) {
	if id < 0 {
		return User{}, fmt.Errorf("error while creating user: invalid id")
	}
	err := ValidateLogin(login)
	if err != nil {
		return User{}, fmt.Errorf("error while creating user: %v", err)
	}
	err = ValidatePassword(password)
	if err != nil {
		return User{}, fmt.Errorf("error while creating user: %v", err)
	}
	password = EncryptPasswordSHA256(password)

	user := config.User{
		Id:       id,
		Login:    login,
		Password: password,
	}

	return User{User: user}, nil
}

func InitUserMap() map[int]config.User {
	users := make(map[int]config.User)

	user, err := CreateUser(1, "heckra@example.com", "StrongPass1!")
	if err != nil {
		fmt.Println("Error creating user 1:", err)
		return nil
	}
	users[user.User.Id] = user.User

	user, err = CreateUser(2, "kostritskoy@example.com", "StrongPass2!")
	if err != nil {
		fmt.Println("Error creating user 2:", err)
		return nil
	}
	users[user.User.Id] = user.User

	user, err = CreateUser(3, "eva@example.com", "StrongPass3!")
	if err != nil {
		fmt.Println("Error creating user 3:", err)
		return nil
	}
	users[user.User.Id] = user.User

	user, err = CreateUser(4, "smart_girl@example.com", "StrongPass4!")
	if err != nil {
		fmt.Println("Error creating user 4:", err)
		return nil
	}
	users[user.User.Id] = user.User

	user, err = CreateUser(5, "cat@example.com", "StrongPass5!")
	if err != nil {
		fmt.Println("Error creating user 5:", err)
		return nil
	}
	users[user.User.Id] = user.User

	return users
}
