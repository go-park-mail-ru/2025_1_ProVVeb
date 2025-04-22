package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type contextKey string

const userIDKey contextKey = "userID"

func PanicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, "Internal server error", 500)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func BodySizeLimitMiddleware(limit int64) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}

func AuthWithCSRFMiddleware(tokenValidator *JwtToken, u *SessionHandler) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionCookie, err := r.Cookie("session_id")
			if err != nil {
				fmt.Println("no cookie for", r.URL.Path)
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			sessionID := sessionCookie.Value
			valueID, err := u.LoginUC.GetSession(sessionID)
			if err != nil {
				fmt.Println("no auth at", r.URL.Path)
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			userID, err := strconv.ParseUint(valueID, 10, 32)
			if err != nil {
				fmt.Println("invalid session userID at", r.URL.Path)
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			sess := &Session{
				ID:     sessionID,
				UserID: uint32(userID),
			}

			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete {
				token := r.Header.Get("csrf-token")
				if token == "" {
					http.Error(w, "Missing CSRF token", http.StatusForbidden)
					return
				}

				valid, err := tokenValidator.CheckJwtToken(sess, token)
				if err != nil || !valid {
					http.Error(w, "Invalid CSRF token", http.StatusForbidden)
					return
				}
			}

			ctx := context.WithValue(r.Context(), userIDKey, uint32(userID))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
