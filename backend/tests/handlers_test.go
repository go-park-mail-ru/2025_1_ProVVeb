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

func TestGetProfile(t *testing.T) {
	tests := []struct {
		id            string
		expectedCode  int
		expectedBody  string
		expectedError string
	}{
		{
			id:           "1",
			expectedCode: http.StatusOK,
			expectedBody: `{"profileId":1,"firstName":"Лиза","lastName":"Тимофеева","height":180,"Birthday":{"year":1990,"month":5,"day":15},"avatar":"http://213.219.214.83:8080/static/avatars/liza.png","card":"http://213.219.214.83:8080/static/cards/liza.png","description":"Специалист по IT","location":"New York","interests":["Technology","Reading","Traveling"],"likedBy":[2,3,4],"Preferences":{"preferencesId":1,"interests":["Music","Movies","Sports"],"location":"New York","Age":{"from":18,"to":35}}}`,
		},
		{
			id:           "invalid_id",
			expectedCode: http.StatusBadRequest,
			expectedBody: "Invalid user ID\n",
		},
		{
			id:           "9999",
			expectedCode: http.StatusNotFound,
			expectedBody: "Profile not found\n",
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

			body := w.Body.String()
			body = strings.TrimSpace(body)
			if w.Code == http.StatusOK && body != tt.expectedBody {
				t.Errorf("expected body %s, got %s", tt.expectedBody, body)
			}
		})
	}
}

func TestCreateUser(t *testing.T) {
	tests := []struct {
		user          config.User
		expectedCode  int
		expectedBody  string
		expectedError string
	}{
		{
			user: config.User{
				Login:    "testLogin",
				Password: "validpassword123",
			},
			expectedCode: http.StatusCreated,
			expectedBody: `{"message":"user created"}`,
		},
		{
			user: config.User{
				Login:    "invalid-login",
				Password: "validpassword#123",
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"message":"invalid email or password"}`,
		},
		{
			user: config.User{
				Login:    "evaecom",
				Password: "StrongPass3",
			},
			expectedCode: http.StatusConflict,
			expectedBody: `{"message":"user already exists"}`,
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

			body := w.Body.String()
			body = strings.TrimSpace(body)
			if body != tt.expectedBody {
				t.Errorf("expected body !%s!, got !%s!", tt.expectedBody, body)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		id           string
		expectedCode int
		expectedBody string
	}{
		{
			id:           "1",
			expectedCode: http.StatusOK,
			expectedBody: `{"message":"User with ID 1 deleted"}`,
		},
		{
			id:           "9999",
			expectedCode: http.StatusNotFound,
			expectedBody: `{"message":"User not found"}`,
		},
		{
			id:           "abc",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"message":"Invalid user ID"}`,
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

			body := w.Body.String()
			body = strings.TrimSpace(body)
			if body != tt.expectedBody {
				t.Errorf("expected body %s, got %s", tt.expectedBody, body)
			}
		})
	}
}

func TestLoginUser(t *testing.T) {
	tests := []struct {
		emai         string
		password     string
		expectedCode int
		expectedBody string
	}{
		{
			emai:         "evaecom",
			password:     "StrongPass3",
			expectedCode: http.StatusOK,
			expectedBody: `{"message":"Logged in","id":3}`,
		},
		{
			emai:         "invalid_id",
			password:     "validpassword#123",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"message":"Invalid login or password"}`,
		},
		{
			emai:         "heckra@example.com",
			password:     "wrongpassword",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"message":"Invalid login or password"}`,
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

			body := w.Body.String()
			body = strings.TrimSpace(body)
			if body != tt.expectedBody {
				t.Errorf("expected body !%s!, got !%s!", tt.expectedBody, body)
			}
		})
	}
}

func TestCheckSession(t *testing.T) {
	tests := []struct {
		sessionID    string
		expectedCode int
		expectedBody string
	}{
		{
			sessionID:    handlers.Testapi.Sessions[3],
			expectedCode: http.StatusOK,
			expectedBody: `{"message":"Logged in","inSession":true,"id":3}`,
		},
		{
			sessionID:    "invalidsessionid",
			expectedCode: http.StatusOK,
			expectedBody: `{"message":"Session not found","inSession":false}`,
		},
		{
			sessionID:    "",
			expectedCode: http.StatusOK,
			expectedBody: `{"message":"Session not found","inSession":false}`,
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

			body := w.Body.String()
			body = strings.TrimSpace(body)
			if body != tt.expectedBody {
				t.Errorf("expected body !%s!, got !%s!", tt.expectedBody, body)
			}
		})
	}
}

func TestLogoutUser(t *testing.T) {
	tests := []struct {
		sessionID    string
		expectedCode int
		expectedBody string
	}{
		{
			sessionID:    handlers.Testapi.Sessions[3],
			expectedCode: http.StatusOK,
			expectedBody: `{"message":"Logged out"}`,
		},
		{
			sessionID:    "invalidsessionid",
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"message":"Session not found"}`,
		},
		{
			sessionID:    "",
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"message":"Session not found"}`,
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

			body := w.Body.String()
			body = strings.TrimSpace(body)
			if body != tt.expectedBody {
				t.Errorf("expected body !%s!, got !%s!", tt.expectedBody, body)
			}
		})
	}
}
