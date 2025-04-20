package repository

import (
	"crypto/sha256"
	"fmt"
)

type PasswordHasher interface {
	Hash(password string) string
	Compare(hashedPassword, login, password string) bool
}

type PassHasher struct{}

func NewPassHasher() (*PassHasher, error) {
	return &PassHasher{}, nil
}

func (ph *PassHasher) Hash(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (ph *PassHasher) Compare(hashedPassword, salt, inputPassword string) bool {
	return hashedPassword == ph.Hash(salt+"_"+inputPassword)
}
