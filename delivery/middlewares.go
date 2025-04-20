package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type contextKey string

const userIDKey contextKey = "userID"

func AdminAuthMiddleware(u *SessionHandler) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := r.Cookie("session_id")

			if err != nil {
				fmt.Println("no cookie for ", r.URL.Path)
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			userID, err := u.LoginUC.GetSession(session.Value)
			if err != nil {
				fmt.Println("no auth at", r.URL.Path)
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

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
