package model

import (
	"errors"
	"time"
)

var PageSize = 10
var SearchLimit = 20

type Preference struct {
	Description string `yaml:"preference_description" json:"preference_description"`
	Value       string `yaml:"preference_value" json:"preference_value"`
}

type Profile struct {
	ProfileId   int          `yaml:"profileId" json:"profileId"`
	FirstName   string       `yaml:"firstName" json:"firstName"`
	LastName    string       `yaml:"lastName" json:"lastName"`
	IsMale      bool         `yaml:"isMale" json:"isMale"`
	Goal        int          `yaml:"goal" json:"goal"`
	Height      int          `yaml:"height" json:"height"`
	Birthday    time.Time    `yaml:"birthday" json:"birthday"`
	Description string       `yaml:"description" json:"description"`
	Location    string       `yaml:"location" json:"location"`
	Interests   []string     `yaml:"interests" json:"interests"`
	LikedBy     []int        `yaml:"likedBy" json:"likedBy"`
	Preferences []Preference `yaml:"preferences" json:"preferences"`
	Parameters  []Preference `yaml:"parameters" json:"parameters"`
	Photos      []string     `yaml:"photos" json:"photos"`
	Premium     struct {
		Status bool `yaml:"Status" json:"Status"`
		Border int  `yaml:"Border" json:"Border"`
	}
}

type SearchProfileRequest struct {
	Input       string       `json:"input"`
	IsMale      string       `json:"isMale"`
	AgeMin      int          `json:"ageMin"`
	AgeMax      int          `json:"ageMax"`
	HeightMin   int          `json:"heightMin"`
	HeightMax   int          `json:"heightMax"`
	Goal        int          `yaml:"goal" json:"goal"`
	Preferences []Preference `yaml:"preferences" json:"preferences"`
	Country     string       `json:"country"`
	City        string       `json:"city"`
}

type FoundProfile struct {
	IDUser   int    `json:"idUser"`
	FirstImg string `json:"firstImgSrc"`
	Fullname string `json:"fullname"`
	Age      int    `json:"age"`
	Goal     int    `yaml:"goal" json:"goal"`
}

type ProfileStats struct {
	LikesGiven         int `json:"likesGiven"`
	LikesReceived      int `json:"likesReceived"`
	Matches            int `json:"matches"`
	ComplaintsMade     int `json:"complaintsMade"`
	ComplaintsReceived int `json:"complaintsReceived"`
	MessagesSent       int `json:"messagesSent"`
	ChatCount          int `json:"chatCount"`
}

var (
	ErrInvalidUserRepoConfig = errors.New("invalid user repository config")
	ErrProfileNotFound       = errors.New("profile not found")
	ErrInvalidProfile        = errors.New("invalid profile")
	ErrDeleteProfile         = errors.New("failed to delete profile")
)
