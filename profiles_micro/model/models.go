package model

import (
	"errors"
	"time"
)

var PageSize = 10

type Preference struct {
	Description string `yaml:"preference_description" json:"preference_description"`
	Value       string `yaml:"preference_value" json:"preference_value"`
}

type Profile struct {
	ProfileId   int          `yaml:"profileId" json:"profileId"`
	FirstName   string       `yaml:"firstName" json:"firstName"`
	LastName    string       `yaml:"lastName" json:"lastName"`
	IsMale      bool         `yaml:"isMale" json:"isMale"`
	Height      int          `yaml:"height" json:"height"`
	Birthday    time.Time    `yaml:"birthday" json:"birthday"`
	Description string       `yaml:"description" json:"description"`
	Location    string       `yaml:"location" json:"location"`
	Interests   []string     `yaml:"interests" json:"interests"`
	LikedBy     []int        `yaml:"likedBy" json:"likedBy"`
	Preferences []Preference `yaml:"preferences" json:"preferences"`
	Photos      []string     `yaml:"photos" json:"photos"`
}

var (
	ErrInvalidUserRepoConfig = errors.New("invalid user repository config")
)
