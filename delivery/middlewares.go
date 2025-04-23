package handlers

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/sirupsen/logrus"
)

type contextKey string

const userIDKey contextKey = "userID"

func PanicMiddleware(logger logger.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.WithFields(&logrus.Fields{
						"method": r.Method,
						"path":   r.URL.Path,
						"error":  err,
						"stack":  string(debug.Stack()),
					}).Error("recovered from panic")

					http.Error(w, "Internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func BodySizeLimitMiddleware(limit int64) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}

func AuthWithCSRFMiddleware(tokenValidator *repository.JwtToken, u *SessionHandler) mux.MiddlewareFunc {
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

			sess := &repository.Session{
				ID:     sessionID,
				UserID: uint32(userID),
			}

			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete {
				csrfCookie, err := r.Cookie("csrf_token")
				if err != nil {
					http.Error(w, "Missing CSRF token", http.StatusForbidden)
					return
				}

				token := csrfCookie.Value

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

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), "request_id", requestID)
		w.Header().Set("X-Request-ID", requestID)

		if logger, ok := r.Context().Value("logger").(*logger.LogrusLogger); ok {
			logger.WithFields(&logrus.Fields{
				"request_id": requestID,
				"method":     r.Method,
				"path":       r.URL.Path,
			}).Debug("request started")
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AccessLogMiddleware(logger logger.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lrw := newLoggingResponseWriter(w)

			requestID := ""
			if ctxVal := r.Context().Value("request_id"); ctxVal != nil {
				requestID = ctxVal.(string)
			}

			next.ServeHTTP(lrw, r)

			logger.WithFields(&logrus.Fields{
				"method":      r.Method,
				"path":        r.URL.Path,
				"remote_addr": r.RemoteAddr,
				"user_agent":  r.UserAgent(),
				"request_id":  requestID,
				"status":      lrw.statusCode,
				"duration":    time.Since(start).String(),
			}).Info("request completed")
		})
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
