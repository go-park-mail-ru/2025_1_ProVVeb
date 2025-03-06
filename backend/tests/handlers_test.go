package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/backend/handlers"
)

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
			expectedBody: `{
				"ProfileId": 1,
				"FirstName": "Xhr",
				"LastName": "Timofeev",
				"Height": 180,
				"Birthday": {
					"Year": 1990,
					"Month": 5,
					"Day": 15
				},
				"Avatar": "",
				"Description": "A tech enthusiast.",
				"Location": "New York",
				"Interests": ["Technology", "Reading", "Traveling"],
				"LikedBy": [2, 3, 4],
				"Preferences": {
					"PreferencesId": 1,
					"Interests": ["Music", "Movies", "Sports"],
					"Location": "New York",
					"Age": {
						"From": 18,
						"To": 35
					}
				}
			}`,
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
			r := httptest.NewRequest("GET", "/profile/"+tt.id, nil)
			w := httptest.NewRecorder()

			h := &handlers.GetHandler{}
			h.GetProfile(w, r)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
			}

			body := w.Body.String()
			if body != tt.expectedBody {
				t.Errorf("expected body %s, got %s", tt.expectedBody, body)
			}
		})
	}
}
