//go:generate easyjson -all  models.go

package model

import (
	"encoding/json"
	"errors"
	"time"
)

var MinPasswordLength = 8
var MaxPasswordLength = 64
var MinLoginLength = 7
var MaxLoginLength = 25

var MaxFileSize int64 = 10 << 20

const Megabyte int = 1 << 23
const MaxQuerySizeStr int = 5
const MaxQuerySizePhoto int = 15 * 6

var MaxProfileViewsWithoutSub = 5

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
	ErrUserGetParamsUC       = errors.New("failed to get user params")
)

//easyjson:json
type Preference struct {
	Description string `yaml:"preference_description" json:"preference_description"`
	Value       string `yaml:"preference_value" json:"preference_value"`
}

//easyjson:json
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
	Premium     Premium      `yaml:"Premium" json:"Premium"`
}

//easyjson:json
type ProfileIsAdmin struct {
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
	Premium     Premium      `yaml:"Premium" json:"Premium"`
	IsAdmin     bool         `yaml:"isAdmin" json:"isAdmin"`
}

//easyjson:json
type Premium struct {
	Status bool  `yaml:"Status" json:"Status"`
	Border int32 `yaml:"Border" json:"Border"`
}

//easyjson:json
type Session struct {
	SessionId string        `yaml:"sessionId" json:"sessionId"`
	UserId    int           `yaml:"userId" json:"userId"`
	Expires   time.Duration `yaml:"expires" json:"expires"`
}

//easyjson:json
type Cookie struct {
	Name     string    `yaml:"name" json:"name"`
	Value    string    `yaml:"value" json:"value"`
	Expires  time.Time `yaml:"expires" json:"expires"`
	HttpOnly bool      `yaml:"httpOnly" json:"httpOnly"`
	Secure   bool      `yaml:"secure" json:"secure"`
	Path     string    `yaml:"path" json:"path"`
}

//easyjson:json
type Query struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	MinScore    int    `yaml:"minScore" json:"minScore"`
	MaxScore    int    `yaml:"maxScore" json:"maxScore"`
}

//easyjson:json
type QueryForUser struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	MinScore    int    `yaml:"minScore" json:"minScore"`
	MaxScore    int    `yaml:"maxScore" json:"maxScore"`
	Score       int    `yaml:"score" json:"score"`
	Answer      string `yaml:"answer" json:"answer"`
}

//easyjson:json
type UsersForQuery struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	MinScore    int    `yaml:"minScore" json:"minScore"`
	MaxScore    int    `yaml:"maxScore" json:"maxScore"`
	Login       string `yaml:"login" json:"login"`
	Answer      string `yaml:"answer" json:"answer"`
	Score       int    `yaml:"score" json:"score"`
}

//easyjson:json
type AnswersForQuery struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	MinScore    int    `yaml:"minScore" json:"minScore"`
	MaxScore    int    `yaml:"maxScore" json:"maxScore"`
	Login       string `yaml:"login" json:"login"`
	Answer      string `yaml:"answer" json:"answer"`
	Score       int    `yaml:"score" json:"score"`
	UserId      int    `yaml:"userId" json:"userId"`
}

//easyjson:json
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

//easyjson:json
type Message struct {
	MessageID int       `yaml:"messageid" json:"messageid"`
	SenderID  int       `yaml:"senderid" json:"senderid"`
	Text      string    `yaml:"text" json:"text"`
	Status    int       `yaml:"status" json:"status"`
	CreatedAt time.Time `yaml:"createdAt" json:"createdAt"`
}

//easyjson:json
type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

//easyjson:json
type CreatePayload struct {
	ChatID  int    `json:"chat_id"`
	UserID  int    `json:"user_id"`
	Content string `json:"content"`
}

//easyjson:json
type DeletePayload struct {
	ChatID    int `json:"chat_id"`
	MessageID int `json:"message_id"`
}

//easyjson:json
type ReadPayload struct {
	ChatID int `json:"chat_id"`
}

//easyjson:json
type ChatNotificationsPayload struct {
	ChatID int `json:"chat_id"`
}

//easyjson:json
type DeleteNotifPayload struct {
	NotifID int `json:"notif_id"`
}

//easyjson:json
type FlowersPayload struct {
	UserID int `json:"user_id"`
}

//easyjson:json
type ReadNotifPayload struct {
	NotifType string `json:"notif_type"`
}

//easyjson:json
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

//easyjson:json
type ComplaintStats struct {
	Total          int       `json:"total_complaints"`
	Rejected       int       `json:"rejected"`
	Pending        int       `json:"pending"`
	Approved       int       `json:"approved"`
	Closed         int       `json:"closed"`
	TotalBy        int       `json:"total_complainants"`
	TotalOn        int       `json:"total_reported"`
	FirstComplaint time.Time `json:"first_complaint"`
	LastComplaint  time.Time `json:"last_complaint"`
}

//easyjson:json
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

//easyjson:json
type FoundProfile struct {
	IDUser   int    `json:"idUser"`
	FirstImg string `json:"firstImgSrc"`
	Fullname string `json:"fullname"`
	Age      int    `json:"age"`
	Goal     int    `yaml:"goal" json:"goal"`
}

//easyjson:json
type Notification struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

//easyjson:json
type NotificationSend struct {
	NotificationID int    `yaml:"notificationID" json:"notificationID"`
	Read           int    `yaml:"read" json:"read"`
	NotifType      string `yaml:"type" json:"type"`
	Content        string `yaml:"content" json:"content"`
}

//easyjson:json
type QueryStats struct {
	TotalAnswers int     `yaml:"TotalAnswers" json:"TotalAnswers"`
	AverageScore float64 `yaml:"AverageScore" json:"AverageScore"`
	MinScore     int     `yaml:"MinScore" json:"MinScore"`
	MaxScore     int     `yaml:"MaxScore" json:"MaxScore"`
}

//easyjson:json
type CreateChatRequest struct {
	FristID  int `json:"firstID"`
	SecondID int `json:"secondID"`
}

//easyjson:json
type DeleteChatRequest struct {
	FristID  int `json:"firstID"`
	SecondID int `json:"secondID"`
}

//easyjson:json
type SetLike struct {
	LikeFrom int `json:"likeFrom"`
	LikeTo   int `json:"likeTo"`
	Status   int `json:"status"`
}

//easyjson:json
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

//easyjson:json
type SignUpRequest struct {
	User    User    `json:"user"`
	Profile Profile `json:"profile"`
}

//easyjson:json
type HandleComplaint struct {
	Complaint_id int `json:"complaint_id"`
	NewStatus    int `json:"new_status"`
}

//easyjson:json
type DeleteComlaint struct {
	Complaint_id int `json:"complaint_id"`
}

//easyjson:json
type FindComplaint struct {
	Complaint_by   int    `json:"complaint_by"`
	Name_by        string `json:"name_by"`
	Complaint_on   int    `json:"complaint_on"`
	Name_on        string `json:"name_on"`
	Complaint_type string `json:"complaint_type"`
	Status         int    `json:"status"`
}

//easyjson:json
type TimeConstraints struct {
	TimeFrom time.Time `json:"time_from"`
	TimeTo   time.Time `json:"time_to"`
}

//easyjson:json
type RawTimeConstraints struct {
	TimeFrom *string `json:"time_from"`
	TimeTo   *string `json:"time_to"`
}

//easyjson:json
type ChangeBorderRequest struct {
	NewBorder int `json:"new_border"`
}

//easyjson:json
type AddSubRequet struct {
	Label string `json:"label"`
}

//easyjson:json
type CreateComplaintRequest struct {
	Complaint_type string `json:"firstID"`
	Complaint_text string `json:"complaint_text"`
	Complaint_on   string `json:"complaint_on"`
}

//easyjson:json
type GetAnswerStatistics struct {
	Query_name string `json:"query_name"`
}

//easyjson:json
type DeleteQueryRequest struct {
	Query_name string `json:"query_name"`
	User_id    int    `json:"user_id"`
}

//easyjson:json
type FindQueryRequest struct {
	Name     string `json:"name"`
	Query_id int    `json:"query_id"`
}

//easyjson:json
type SessionCheckResponse struct {
	Message   string `json:"message"`
	InSession bool   `json:"inSession"`
}

//easyjson:json
type SessionCheckSuccessResponse struct {
	Message   string `json:"message"`
	InSession bool   `json:"inSession"`
	UserId    int    `json:"id"`
}

//easyjson:json
type ErrorResponse struct {
	Message string `json:"message"`
}

//easyjson:json
type ComplaintsResponse struct {
	Complaints []ComplaintWithLogins `json:"complaints"`
}

//easyjson:json
type LoginResponse struct {
	Message string `json:"message"`
	UserID  int    `json:"user_id"`
}

//easyjson:json
type UploadResponse struct {
	Message          string   `json:"message"`
	SucessfulUploads []string `json:"sucessful_uploads"`
}

//easyjson:json
type ProfileResponse struct {
	Profiles []Profile `json:"profiles"`
}

//easyjson:json
type FoundProfileResponse struct {
	Profiles []FoundProfile `json:"profiles"`
}

//easyjson:json
type ChatsResponse struct {
	Chats []Chat `json:"chats"`
}

//easyjson:json
type AnswersResponse struct {
	Answers []UsersForQuery `json:"answers"`
}

//easyjson:json
type AnswersForResponse struct {
	Answers []AnswersForQuery `json:"answers"`
}

//easyjson:json
type QueryResponse struct {
	Queries []QueryForUser `json:"queryis"`
}

//easyjson:json
type QuerResponse struct {
	Queries []Query `json:"queryis"`
}

//easyjson:json
type User struct {
	UserId   int    `yaml:"id" json:"id"`
	Login    string `yaml:"login" json:"login"`
	Password string `yaml:"password" json:"password"`
	Email    string `yaml:"email" json:"email"`
	Phone    string `yaml:"phone" json:"phone"`
	Status   int    `yaml:"status" json:"status"`
}

//easyjson:json
type UserAnswer struct {
	Name   string `json:"name"`
	Score  int32  `json:"score"`
	Answer string `json:"answer"`
}

//easyjson:json
type ProfileStats struct {
	LikesGiven         int `json:"likesGiven"`
	LikesReceived      int `json:"likesReceived"`
	Matches            int `json:"matches"`
	ComplaintsMade     int `json:"complaintsMade"`
	ComplaintsReceived int `json:"complaintsReceived"`
	MessagesSent       int `json:"messagesSent"`
	ChatCount          int `json:"chatCount"`
}
