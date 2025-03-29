package config

import "time"

var MinPasswordLength = 8
var MaxPasswordLength = 64
var MinLoginLength = 7
var MaxLoginLength = 15

type User struct {
	UserId   int    `yaml:"id" json:"id"`
	Login    string `yaml:"login" json:"login"`
	Password string `yaml:"password" json:"password"`
	Email    string `yaml:"email" json:"email"`
	Phone    string `yaml:"phone" json:"phone"`
	Status   int    `yaml:"status" json:"status"`
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type Profile struct {
	ProfileId   int       `yaml:"profileId" json:"profileId"`
	FirstName   string    `yaml:"firstName" json:"firstName"`
	LastName    string    `yaml:"lastName" json:"lastName"`
	IsMale      bool      `yaml:"isMale" json:"isMale"`
	Height      int       `yaml:"height" json:"height"`
	Birthday    time.Time `yaml:"birthday" json:"birthday"`
	Avatar      string    `yaml:"avatar" json:"avatar"`
	Card        string    `yaml:"card" json:"card"`
	Description string    `yaml:"description" json:"description"`
	Location    string    `yaml:"location" json:"location"`
	Interests   []string  `yaml:"interests" json:"interests"`
	LikedBy     []int     `yaml:"likedBy" json:"likedBy"`
	Preferences []string  `yaml:"preferences" json:"preferences"`
}
