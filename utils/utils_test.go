package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateLogin(t *testing.T) {
	validLogins := []string{"user_la1", "User_123", "ab_bcbc", "Zaaaaaa"}
	for _, login := range validLogins {
		assert.NoError(t, ValidateLogin(login), "login should be valid: "+login)
	}

	assert.Error(t, ValidateLogin("a"), "too short login")
	assert.Error(t, ValidateLogin("thisloginistoolongtobevalidbecauseoflength12345"), "too long login")

	assert.Error(t, ValidateLogin("1user"), "login starts with digit")
	assert.Error(t, ValidateLogin("user!"), "login contains invalid char")
	assert.Error(t, ValidateLogin("user space"), "login contains space")
}

func TestValidatePassword(t *testing.T) {
	validPasswords := []string{"Password123", "abcDEF456", "12345678", "a1b2c3d4"}
	for _, pass := range validPasswords {
		assert.NoError(t, ValidatePassword(pass), "password should be valid: "+pass)
	}

}

func TestEncryptPasswordSHA256(t *testing.T) {
	pass := "password"
	hash := EncryptPasswordSHA256(pass)
	assert.Len(t, hash, 64)
}

func TestCreateUser(t *testing.T) {
	u, err := CreateUser(1, "validUser", "Password123")
	assert.NoError(t, err)
	assert.Equal(t, 1, u.User.UserId)
	assert.Equal(t, "validUser", u.User.Login)
	assert.NotEmpty(t, u.User.Password)

	_, err = CreateUser(-1, "validUser", "Password123")
	assert.Error(t, err)

	// Ошибка из-за логина
	_, err = CreateUser(1, "1invalid", "Password123")
	assert.Error(t, err)

	// Ошибка из-за пароля
	_, err = CreateUser(1, "validUser", "bad pass")
	assert.Error(t, err)
}
