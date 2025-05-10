package model

import (
	"encoding/json"
	"errors"
	"time"
)

var MinPasswordLength = 8
var MaxPasswordLength = 64
var MinLoginLength = 7
var MaxLoginLength = 15

var PageSize = 10
var MaxFileSize int64 = 10 << 20

const Megabyte int = 1 << 23
const MaxQuerySizeStr int = 5
const MaxQuerySizePhoto int = 15 * 6

var Key string = "Hello"

// regexps
var (
	ReStartsWithLetter             = `^[a-zA-Z]`
	ReContainsLettersDigitsSymbols = `^[a-zA-Z0-9._-]+$`
)

// ideas for future
// password must contain at least one digit
// password must contain only letters and digits
// password must contain at least one special character
// password must not contain invalid characters

// errors
var (
	ErrInvalidLogin          = errors.New("invalid login")
	ErrInvalidLoginSize      = errors.New("invalid login size")
	ErrInvalidPassword       = errors.New("invalid password")
	ErrInvalidPasswordSize   = errors.New("invalid password size")
	ErrSessionNotFound       = errors.New("session not found")
	ErrInvalidSession        = errors.New("invalid session")
	ErrInvalidUserRepoConfig = errors.New("invalid user repository config")
	ErrGetSession            = errors.New("failed to get session")
	ErrStoreSession          = errors.New("failed to store session")
	ErrInvalidSessionId      = errors.New("invalid session id")
	ErrDeleteSession         = errors.New("failed to delete session")
	ErrProfileNotFound       = errors.New("profile not found")
	ErrDeleteUser            = errors.New("failed to delete user")
	ErrDeleteProfile         = errors.New("failed to delete profile")
	ErrUserCheckSessionUC    = errors.New("failed user check session")
	ErrDeleteStaticUC        = errors.New("failed to delete static")
	ErrUserDeleteUC          = errors.New("failed to delete user")
	ErrGetProfileMatchesUC   = errors.New("failed to get profile matches")
	ErrGetProfileUC          = errors.New("failed to get profile")
	ErrGetUserPhotoUC        = errors.New("failed to get user photo")
	ErrGetProfilesForUserUC  = errors.New("failed to get profiles for user")
	ErrProfileSetLikeUC      = errors.New("failed to set like")
	ErrUserLogInUC           = errors.New("failed to log in user")
	ErrUserLogOutUC          = errors.New("failed to log out user")
	ErrUserSignUpUC          = errors.New("failed to sign up user")
	ErrProfileUpdateUC       = errors.New("failed to update profile")
	ErrStaticUploadUC        = errors.New("failed to upload static")
	ErrGetActiveQueriesUC    = errors.New("failed to get active queries")
)

type User struct {
	UserId   int    `yaml:"id" json:"id"`
	Login    string `yaml:"login" json:"login"`
	Password string `yaml:"password" json:"password"`
	Email    string `yaml:"email" json:"email"`
	Phone    string `yaml:"phone" json:"phone"`
	Status   int    `yaml:"status" json:"status"`
}

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

type Session struct {
	SessionId string        `yaml:"sessionId" json:"sessionId"`
	UserId    int           `yaml:"userId" json:"userId"`
	Expires   time.Duration `yaml:"expires" json:"expires"`
}

type Cookie struct {
	Name     string    `yaml:"name" json:"name"`
	Value    string    `yaml:"value" json:"value"`
	Expires  time.Time `yaml:"expires" json:"expires"`
	HttpOnly bool      `yaml:"httpOnly" json:"httpOnly"`
	Secure   bool      `yaml:"secure" json:"secure"`
	Path     string    `yaml:"path" json:"path"`
}

type Query struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	MinScore    int    `yaml:"minScore" json:"minScore"`
	MaxScore    int    `yaml:"maxScore" json:"maxScore"`
}

type QueryForUser struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	MinScore    int    `yaml:"minScore" json:"minScore"`
	MaxScore    int    `yaml:"maxScore" json:"maxScore"`
	Score       int    `yaml:"score" json:"score"`
	Answer      string `yaml:"answer" json:"answer"`
}

type UsersForQuery struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	MinScore    int    `yaml:"minScore" json:"minScore"`
	MaxScore    int    `yaml:"maxScore" json:"maxScore"`
	Login       string `yaml:"login" json:"login"`
	Answer      string `yaml:"answer" json:"answer"`
	Score       int    `yaml:"score" json:"score"`
}

type Chat struct {
	ProfileId          int    `yaml:"profileId" json:"profileId"`
	ChatId             int    `yaml:"chatId" json:"chatId"`
	ProfileName        string `yaml:"profileName" json:"profileName"`
	ProfilePicture     string `yaml:"profilePicture" json:"profilePicture"`
	ProfileDescription string `yaml:"profileDescription" json:"profileDescription"`
	LastMessage        string `yaml:"lastMessage" json:"lastMessage"`
	IsRead             bool   `yaml:"isRead" json:"isRead"`
	IsSelf             bool   `yaml:"isSelf" json:"isSelf"`
}

type Message struct {
	MessageID int       `yaml:"messageid" json:"messageid"`
	SenderID  int       `yaml:"senderid" json:"senderid"`
	Text      string    `yaml:"text" json:"text"`
	Status    int       `yaml:"status" json:"status"`
	CreatedAt time.Time `yaml:"createdAt" json:"createdAt"`
}

type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type CreatePayload struct {
	ChatID  int    `json:"chat_id"`
	UserID  int    `json:"user_id"`
	Content string `json:"content"`
}

type DeletePayload struct {
	ChatID    int `json:"chat_id"`
	MessageID int `json:"message_id"`
}

type ReadPayload struct {
	ChatID int `json:"chat_id"`
}

type ChatNotificationsPayload struct {
	ChatID int `json:"chat_id"`
}

type ComplaintWithLogins struct {
	ComplaintID   int64      `json:"complaint_id"`
	ComplaintBy   string     `json:"complaint_by"`
	ComplaintOn   string     `json:"complaint_on"`
	ComplaintType int64      `json:"complaint_type"`
	TypeDesc      string     `json:"type_description"`
	Text          string     `json:"complaint_text"`
	Status        int        `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	ClosedAt      *time.Time `json:"closed_at"`
}

type SearchProfileRequest struct {
	Input     string `json:"input"`
	IsMale    string `json:"isMale"`
	AgeMin    int    `json:"ageMin"`
	AgeMax    int    `json:"ageMax"`
	HeightMin int    `json:"heightMin"`
	HeightMax int    `json:"heightMax"`
	Country   string `json:"country"`
	City      string `json:"city"`
}

type FoundProfile struct {
	IDUser   int    `json:"idUser"`
	FirstImg string `json:"firstImgSrc"`
	Fullname string `json:"fullname"`
	Age      int    `json:"age"`
}
