package handlers

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sirupsen/logrus"
)

type contextKey string

const userIDKey contextKey = "userID"

const isPremiumKey contextKey = "isPremium"

const requestIDKey contextKey = "request_id"

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

func AuthWithCSRFMiddleware(tokenValidator *repository.JwtToken, sessionHandler *SessionHandler, userHandler *UserHandler) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionCookie, err := r.Cookie("session_id")
			fmt.Println("Request headers:", r.Header)
			if err != nil {
				fmt.Println("no cookie for", r.URL.Path)
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			sessionID := sessionCookie.Value
			valueID, err := sessionHandler.LoginUC.GetSession(sessionID)
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

			is_premium, _, err := userHandler.GetPremiumUC.GetPremium(int(userID))
			if err != nil {
				fmt.Println("Error getting premium at...", r.URL.Path)
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
			ctx = context.WithValue(ctx, isPremiumKey, is_premium)
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

		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
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
			lrw := NewResponseWriter(w)

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

type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (lrw *ResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := lrw.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("underlying ResponseWriter does not support Hijacker")
}

var (
	httpRequests  *prometheus.CounterVec
	httpDuration  *prometheus.HistogramVec
	syscallReads  *prometheus.CounterVec
	syscallWrites *prometheus.CounterVec
)

type MetricsMiddlewareConfig struct {
	Registry *prometheus.Registry
}

func NewMetricsMiddleware(cfg MetricsMiddlewareConfig) mux.MiddlewareFunc {
	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests",
		},
		[]string{"path", "method", "status"},
	)

	httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: []float64{0.1, 0.5, 1, 2.5, 5},
		},
		[]string{"path", "method"},
	)

	syscallReads = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_syscall_reads_total",
			Help: "Total read syscalls",
		},
		[]string{"pid", "syscalls"},
	)
	syscallWrites = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_syscall_writes_total",
			Help: "Total write syscalls",
		},
		[]string{"pid", "syscalls"},
	)

	cfg.Registry.MustRegister(httpRequests, httpDuration)
	cfg.Registry.MustRegister(syscallReads, syscallWrites)
	cfg.Registry.MustRegister(
		messageChatsCreated,
		messageChatsViews,
		messageChatsDeleted,
		messageSent,
		messageReceived,
		messageNotificationsFetched,

		profileRetrieved,
		profileUpdated,
		photoRemoved,
		likeSet,
		searchPerformed,
		matchesRetrieved,
		photoUploaded,
		profilesListRetrieved,

		loginAttempts,
		sessionChecks,
		logoutAttempts,
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &ResponseWriter{w, http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start).Seconds()

			httpRequests.WithLabelValues(
				GetUrlForMetrics(r.URL.Path),
				r.Method,
				strconv.Itoa(rw.statusCode),
			).Inc()

			httpDuration.WithLabelValues(
				GetUrlForMetrics(r.URL.Path),
				r.Method,
			).Observe(duration)
		})
	}
}

func updateSyscallMetrics() {
	pid := os.Getpid()
	pidStr := strconv.Itoa(pid)

	ioData, err := os.ReadFile(fmt.Sprintf("/proc/%d/io", pid))
	if err != nil {
		log.Printf("Failed to read io data: %v", err)
		return
	}

	reads, writes := parseIOStats(string(ioData))

	syscallReads.WithLabelValues(pidStr, "").Add(float64(reads))
	syscallWrites.WithLabelValues(pidStr, "").Add(float64(writes))
}

func parseIOStats(data string) (reads, writes uint64) {
	scanner := bufio.NewScanner(strings.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "rchar:"):
			parts := strings.Fields(line)
			if len(parts) > 1 {
				reads, _ = strconv.ParseUint(parts[1], 10, 64)
			}
		case strings.HasPrefix(line, "wchar:"):
			parts := strings.Fields(line)
			if len(parts) > 1 {
				writes, _ = strconv.ParseUint(parts[1], 10, 64)
			}
		}
	}
	return reads, writes
}
