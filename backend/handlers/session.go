package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	postgres "github.com/go-park-mail-ru/2025_1_ProVVeb/backend/database_function/postgres/queries"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/backend/database_function/redis"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/backend/utils"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/config"
	"github.com/jackc/pgx/v5"
)

var muSessions = &sync.Mutex{}

type SessionHandler struct {
	DB          *pgx.Conn
	RedisClient *redis.RedisClient
}

func RandStringRunes(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (u *SessionHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var gotData struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&gotData); err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid JSON data"})
		return
	}

	if (utils.ValidateLogin(gotData.Login) != nil) || (utils.ValidatePassword(gotData.Password) != nil) {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid login or password"})
		return
	}

	login, password := gotData.Login, gotData.Password

	var foundUser config.User
	foundUser, err := postgres.DBGetUserPostgres(u.DB, login)

	fmt.Println(err)
	if err != nil {
		makeResponse(w, http.StatusNotFound, map[string]string{"message": "No such user"})
		return
	}

	if foundUser.Password != utils.EncryptPasswordSHA256(password) {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid login or password"})
		return
	}

	SID := RandStringRunes(32)
	muSessions.Lock()
	defer muSessions.Unlock()

	err = u.RedisClient.StoreSession(SID, fmt.Sprintf("%d", foundUser.UserId), 72*time.Hour)
	if err != nil {
		makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "Failed to store session"})
		return
	}

	// api.sessions[SID] = foundUser.UserId
	// Testapi.Sessions[foundUser.UserId] = SID // для теста Logout

	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    SID,
		HttpOnly: true,
		Secure:   false,
		Expires:  time.Now().Add(3 * 24 * time.Hour),
		Path:     "/",
	}

	http.SetCookie(w, cookie)

	response := struct {
		Message string `json:"message"`
		UserId  int    `json:"id"`
	}{
		Message: "Logged in",
		UserId:  foundUser.UserId,
	}

	makeResponse(w, http.StatusOK, response)
}

func (u *SessionHandler) CheckSession(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		response := struct {
			Message   string `json:"message"`
			InSession bool   `json:"inSession"`
		}{
			Message:   "No cookies got",
			InSession: false,
		}

		makeResponse(w, http.StatusOK, response)
		return
	}

	muSessions.Lock()
	defer muSessions.Unlock()

	userIdStr, err := u.RedisClient.GetSession(session.Value)
	if err != nil {
		if err.Error() == "redis: nil" {
			response := struct {
				Message   string `json:"message"`
				InSession bool   `json:"inSession"`
			}{
				Message:   "Session not found",
				InSession: false,
			}

			makeResponse(w, http.StatusOK, response)
			return
		}

		makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "Failed to check session"})
		return
	}

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "Invalid session data"})
		return
	}

	response := struct {
		Message   string `json:"message"`
		InSession bool   `json:"inSession"`
		UserId    int    `json:"id"`
	}{
		Message:   "Logged in",
		InSession: true,
		UserId:    userId,
	}

	makeResponse(w, http.StatusOK, response)
}

func (u *SessionHandler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "No cookies got"})
		return
	}

	muSessions.Lock()
	defer muSessions.Unlock()

	err = u.RedisClient.DeleteSession(session.Value)
	if err != nil {
		makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "Failed to delete session"})
		return
	}

	expiredCookie := &http.Cookie{
		Name:     "session_id",
		Value:    "",
		HttpOnly: true,
		Secure:   false,
		Expires:  time.Now().AddDate(-1, 0, 0),
		Path:     "/",
	}

	http.SetCookie(w, expiredCookie)

	makeResponse(w, http.StatusOK, map[string]string{"message": "Logged out"})
}
