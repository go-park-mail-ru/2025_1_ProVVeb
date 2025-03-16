package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/backend/handlers"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/backend/utils"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/config"
	"github.com/gorilla/mux"
)

type Message struct {
	Message string `json:"message"`
}

var testUsers = utils.InitUserMap()

var Se = struct {
	users map[int]config.User
}{
	users: map[int]config.User{
		1: testUsers[1],
	},
}

var api = struct {
	sessions map[string]int
}{sessions: make(map[string]int)}

func TestGetProfilePositive(t *testing.T) {
	tests := []struct {
		id           string
		expectedCode int
		expectedBody config.Profile
	}{
		{
			id:           "1",
			expectedCode: http.StatusOK,
			expectedBody: config.Profile{
				ProfileId: 1,
				FirstName: "Лиза",
				LastName:  "Тимофеева",
				Height:    180,
				Birthday: struct {
					Year  int `yaml:"year" json:"year"`
					Month int `yaml:"month" json:"month"`
					Day   int `yaml:"day" json:"day"`
				}{
					Year:  1990,
					Month: 5,
					Day:   15,
				},
				Avatar:      "http://213.219.214.83:8080/static/avatars/liza.png",
				Card:        "http://213.219.214.83:8080/static/cards/liza.png",
				Description: "Специалист по IT",
				Location:    "New York",
				Interests:   []string{"Technology", "Reading", "Traveling"},
				LikedBy:     []int{2, 3, 4},
				Preferences: struct {
					PreferencesId int      `yaml:"preferencesId" json:"preferencesId"`
					Interests     []string `yaml:"interests" json:"interests"`
					Location      string   `yaml:"location" json:"location"`
					Age           struct {
						From int `yaml:"from" json:"from"`
						To   int `yaml:"to" json:"to"`
					}
				}{
					PreferencesId: 1,
					Interests:     []string{"Music", "Movies", "Sports"},
					Location:      "New York",
					Age: struct {
						From int `yaml:"from" json:"from"`
						To   int `yaml:"to" json:"to"`
					}{
						From: 18,
						To:   35,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("GetProfile ID=%s", tt.id), func(t *testing.T) {
			r := httptest.NewRequest("GET", "/profiles/"+tt.id, nil)

			w := httptest.NewRecorder()

			h := &handlers.GetHandler{}

			router := mux.NewRouter()
			router.HandleFunc("/profiles/{id}", h.GetProfile)

			router.ServeHTTP(w, r)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
			}

			var actualBody config.Profile
			err := json.Unmarshal(w.Body.Bytes(), &actualBody)
			if err != nil {
				t.Fatalf("failed to unmarshal response body: %v", err)
			}

			if !utils.CompareProfiles(actualBody, tt.expectedBody) {
				t.Errorf("expected profile %+v, got %+v", tt.expectedBody, actualBody)
			}
		})
	}
}

func TestGetNegative(t *testing.T) {
	tests := []struct {
		id           string
		expectedCode int
		expectedBody Message
	}{
		{
			id:           "invalid_id",
			expectedCode: http.StatusBadRequest,
			expectedBody: Message{Message: "Invalid user ID"},
		},
		{
			id:           "9999",
			expectedCode: http.StatusNotFound,
			expectedBody: Message{Message: "Profile not found"},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("GetProfile ID=%s", tt.id), func(t *testing.T) {
			r := httptest.NewRequest("GET", "/profiles/"+tt.id, nil)

			w := httptest.NewRecorder()

			h := &handlers.GetHandler{}

			router := mux.NewRouter()
			router.HandleFunc("/profiles/{id}", h.GetProfile)

			router.ServeHTTP(w, r)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
			}

			var actualBody Message
			err := json.Unmarshal(w.Body.Bytes(), &actualBody)
			if err != nil {
				t.Fatalf("failed to unmarshal response body: %v", err)
			}

			if actualBody != tt.expectedBody {
				t.Errorf("expected body %+v, got %+v", tt.expectedBody, actualBody)
			}
		})
	}
}

func TestCreateUser(t *testing.T) {
	tests := []struct {
		user         config.User
		expectedCode int
		expectedBody Message
	}{
		{
			user: config.User{
				Login:    "testLogin",
				Password: "validpassword123",
			},
			expectedCode: http.StatusCreated,
			expectedBody: Message{
				Message: "user created",
			},
		},
		{
			user: config.User{
				Login:    "invalid-login",
				Password: "validpassword#123",
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: Message{
				Message: "invalid email or password",
			},
		},
		{
			user: config.User{
				Login:    "evaecom",
				Password: "StrongPass3",
			},
			expectedCode: http.StatusConflict,
			expectedBody: Message{
				Message: "user already exists",
			},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("CreateUser Login=%s", tt.user.Login), func(t *testing.T) {
			userJSON, err := json.Marshal(tt.user)
			if err != nil {
				t.Fatalf("Failed to marshal user: %v", err)
			}

			r := httptest.NewRequest("POST", "/users", bytes.NewReader(userJSON))
			r.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			h := &handlers.UserHandler{}
			router := mux.NewRouter()
			router.HandleFunc("/users", h.CreateUser)

			router.ServeHTTP(w, r)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
			}

			var actualBody Message
			err = json.Unmarshal(w.Body.Bytes(), &actualBody)
			if err != nil {
				t.Fatalf("Failed to unmarshal response body: %v", err)
			}

			if actualBody != tt.expectedBody {
				t.Errorf("expected body %+v, got %+v", tt.expectedBody, actualBody)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		id           string
		expectedCode int
		expectedBody Message
	}{
		{
			id:           "1",
			expectedCode: http.StatusOK,
			expectedBody: Message{
				Message: "User with ID 1 deleted",
			},
		},
		{
			id:           "9999",
			expectedCode: http.StatusNotFound,
			expectedBody: Message{
				Message: "User not found",
			},
		},
		{
			id:           "abc",
			expectedCode: http.StatusBadRequest,
			expectedBody: Message{
				Message: "Invalid user ID",
			},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("DeleteUser ID=%s", tt.id), func(t *testing.T) {
			if tt.id == "1" {
				testUsers[1] = config.User{
					Login:    "userToDelete@mail.com",
					Password: "password123",
				}
			}

			r := httptest.NewRequest("DELETE", "/users/"+tt.id, nil)
			r = mux.SetURLVars(r, map[string]string{
				"id": tt.id,
			})

			w := httptest.NewRecorder()

			h := &handlers.UserHandler{}
			router := mux.NewRouter()
			router.HandleFunc("/users/{id}", h.DeleteUser)

			router.ServeHTTP(w, r)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
			}

			var actualBody Message
			err := json.Unmarshal(w.Body.Bytes(), &actualBody)
			if err != nil {
				t.Fatalf("Failed to unmarshal response body: %v", err)
			}

			if actualBody != tt.expectedBody {
				t.Errorf("expected body %+v, got %+v", tt.expectedBody, actualBody)
			}
		})
	}
}

func TestLoginUser(t *testing.T) {
	tests := []struct {
		emai         string
		password     string
		expectedCode int
		expectedBody Message
	}{
		{
			emai:         "evaecom",
			password:     "StrongPass3",
			expectedCode: http.StatusOK,
			expectedBody: Message{
				Message: "Logged in",
			},
		},
		{
			emai:         "invalid_id",
			password:     "validpassword#123",
			expectedCode: http.StatusBadRequest,
			expectedBody: Message{
				Message: "Invalid login or password",
			},
		},
		{
			emai:         "heckra@example.com",
			password:     "wrongpassword",
			expectedCode: http.StatusBadRequest,
			expectedBody: Message{
				Message: "Invalid login or password",
			},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Login with login=%s password=%s", tt.emai, tt.password), func(t *testing.T) {
			r := httptest.NewRequest("POST", "/users/login", strings.NewReader(fmt.Sprintf(`{"Login":"%s","Password":"%s"}`, tt.emai, tt.password)))
			w := httptest.NewRecorder()

			h := &handlers.SessionHandler{}
			router := mux.NewRouter()
			router.HandleFunc("/users/login", h.LoginUser)

			router.ServeHTTP(w, r)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
			}

			var actualBody Message
			err := json.Unmarshal(w.Body.Bytes(), &actualBody)
			if err != nil {
				t.Fatalf("Failed to unmarshal response body: %v", err)
			}

			if actualBody != tt.expectedBody {
				t.Errorf("expected body %+v, got %+v", tt.expectedBody, actualBody)
			}
		})
	}
}

func TestCheckSession(t *testing.T) {
	tests := []struct {
		sessionID    string
		expectedCode int
		expectedBody Message
	}{
		{
			sessionID:    handlers.Testapi.Sessions[3],
			expectedCode: http.StatusOK,
			expectedBody: Message{
				Message: "Logged in",
			},
		},
		{
			sessionID:    "invalidsessionid",
			expectedCode: http.StatusOK,
			expectedBody: Message{
				Message: "Session not found",
			},
		},
		{
			sessionID:    "",
			expectedCode: http.StatusOK,
			expectedBody: Message{
				Message: "Session not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("CheckSession with sessionID=%s", tt.sessionID), func(t *testing.T) {
			if tt.sessionID != "" {
				api.sessions[tt.sessionID] = 1
			}
			cookie := &http.Cookie{
				Name:     "session_id",
				Value:    tt.sessionID,
				HttpOnly: true,
				Secure:   false,
				Expires:  time.Now().Add(10 * time.Hour),
			}

			r := httptest.NewRequest("GET", "/users/checkSession", nil)
			r.AddCookie(cookie)
			w := httptest.NewRecorder()

			h := &handlers.SessionHandler{}
			router := mux.NewRouter()
			router.HandleFunc("/users/checkSession", h.CheckSession)

			router.ServeHTTP(w, r)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
			}

			var actualBody Message
			err := json.Unmarshal(w.Body.Bytes(), &actualBody)
			if err != nil {
				t.Fatalf("Failed to unmarshal response body: %v", err)
			}

			if actualBody != tt.expectedBody {
				t.Errorf("expected body %+v, got %+v", tt.expectedBody, actualBody)
			}
		})
	}
}

func TestLogoutUser(t *testing.T) {
	tests := []struct {
		sessionID    string
		expectedCode int
		expectedBody Message
	}{
		{
			sessionID:    handlers.Testapi.Sessions[3],
			expectedCode: http.StatusOK,
			expectedBody: Message{
				Message: "Logged out",
			},
		},
		{
			sessionID:    "invalidsessionid",
			expectedCode: http.StatusUnauthorized,
			expectedBody: Message{
				Message: "Session not found",
			},
		},
		{
			sessionID:    "",
			expectedCode: http.StatusUnauthorized,
			expectedBody: Message{
				Message: "Session not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Logout with sessionID=%s", tt.sessionID), func(t *testing.T) {
			if tt.sessionID != "" {
				api.sessions[tt.sessionID] = 1
			}
			cookie := &http.Cookie{
				Name:     "session_id",
				Value:    tt.sessionID,
				HttpOnly: true,
				Secure:   false,
				Expires:  time.Now().Add(10 * time.Hour),
			}

			r := httptest.NewRequest("POST", "/users/logout", nil)
			r.AddCookie(cookie)
			w := httptest.NewRecorder()

			h := &handlers.SessionHandler{}
			router := mux.NewRouter()
			router.HandleFunc("/users/logout", h.LogoutUser)

			router.ServeHTTP(w, r)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
			}

			var actualBody Message
			err := json.Unmarshal(w.Body.Bytes(), &actualBody)
			if err != nil {
				t.Fatalf("Failed to unmarshal response body: %v", err)
			}

			if actualBody != tt.expectedBody {
				t.Errorf("expected body %+v, got %+v", tt.expectedBody, actualBody)
			}
		})
	}
}
