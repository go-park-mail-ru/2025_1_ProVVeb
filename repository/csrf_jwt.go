package repository

import (
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type JwtTokenizer interface {
	CreateJwtToken(s *Session, tokenExpTime int64) (string, error)
	CheckJwtToken(s *Session, inputToken string) (bool, error)
	ParseSecretGetter(token *jwt.Token) (interface{}, error)
	ExtractSessionFromToken(token string) (*Session, error)
}

type JwtToken struct {
	Secret []byte
}

type Session struct {
	UserID uint32
	ID     string
}

type JwtCsrfClaims struct {
	SessionID string `json:"sid"`
	UserID    uint32 `json:"uid"`
	jwt.StandardClaims
}

func NewJwtToken(secret string) (*JwtToken, error) {
	return &JwtToken{Secret: []byte(secret)}, nil
}

func (tk *JwtToken) CreateJwtToken(s *Session, tokenExpTime int64) (string, error) {
	data := JwtCsrfClaims{
		SessionID: s.ID,
		UserID:    s.UserID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: tokenExpTime,
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	return token.SignedString(tk.Secret)
}

func (tk *JwtToken) ParseSecretGetter(token *jwt.Token) (interface{}, error) {
	method, ok := token.Method.(*jwt.SigningMethodHMAC)
	if !ok || method.Alg() != "HS256" {
		return nil, fmt.Errorf("bad sign method")
	}
	return tk.Secret, nil
}

func (tk *JwtToken) CheckJwtToken(s *Session, inputToken string) (bool, error) {
	payload := &JwtCsrfClaims{}
	_, err := jwt.ParseWithClaims(inputToken, payload, tk.ParseSecretGetter)
	if err != nil {
		return false, fmt.Errorf("cant parse jwt token: %v", err)
	}
	if payload.Valid() != nil {
		return false, fmt.Errorf("invalid jwt token: %v", err)
	}
	return payload.SessionID == s.ID && payload.UserID == s.UserID, nil
}

func (tk *JwtToken) ExtractSessionFromToken(token string) (*Session, error) {
	payload := &JwtCsrfClaims{}
	_, err := jwt.ParseWithClaims(token, payload, tk.ParseSecretGetter)
	if err != nil {
		return nil, fmt.Errorf("can't parse jwt token: %v", err)
	}
	if payload.Valid() != nil {
		return nil, fmt.Errorf("invalid jwt token")
	}
	return &Session{
		ID:     payload.SessionID,
		UserID: payload.UserID,
	}, nil
}
