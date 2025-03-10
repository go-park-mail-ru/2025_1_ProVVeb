package handlers

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/backend/utils"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/config"
)

var Testapi = struct {
	Sessions map[int]string
}{Sessions: make(map[int]string)}

var api = struct {
	sessions map[string]int
}{sessions: make(map[string]int)}

type SessionHandler struct{}

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
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid JSON data"})
	}

	if (utils.ValidateLogin(gotData.Login) != nil) || (utils.ValidatePassword(gotData.Password) != nil) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "invalid login or password"})
		return
	}

	login, password := gotData.Login, gotData.Password

	var foundUser *config.User
	for _, user := range Users {
		if user.Login == login {
			foundUser = &user
			break
		}
	}
	if foundUser == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "no such user"})
		return
	}

	if foundUser.Password != utils.EncryptPasswordSHA256(password) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "invalid login or password"})
		return
	}

	SID := RandStringRunes(32)
	api.sessions[SID] = foundUser.Id
	Testapi.Sessions[foundUser.Id] = SID // для теста Logout

	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    SID,
		HttpOnly: true,
		Secure:   false,
		Expires:  time.Now().Add(10 * time.Hour),
		Path:     "/",
	}

	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)

	response := struct {
		Message string `json:"message"`
		UserId  int    `json:"id"`
	}{
		Message: "Logged in",
		UserId:  foundUser.Id,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (u *SessionHandler) CheckSession(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		w.WriteHeader(http.StatusOK)
		response := struct {
			Message   string `json:"message"`
			InSession bool   `json:"inSession"`
		}{
			Message:   "No cookies got",
			InSession: false,
		}

		json.NewEncoder(w).Encode(response)
		return
	}

	userId, ok := api.sessions[session.Value]
	if !ok {
		w.WriteHeader(http.StatusOK)
		response := struct {
			Message   string `json:"message"`
			InSession bool   `json:"inSession"`
		}{
			Message:   "Session not found",
			InSession: false,
		}

		json.NewEncoder(w).Encode(response)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := struct {
		Message   string `json:"message"`
		InSession bool   `json:"inSession"`
		UserId    int    `json:"id"`
	}{
		Message:   "Logged in",
		InSession: true,
		UserId:    userId,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (u *SessionHandler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "No cookies got"})
		return
	}

	if _, ok := api.sessions[session.Value]; !ok {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Session not found"})
		return
	}

	delete(api.sessions, session.Value)
	delete(Testapi.Sessions, api.sessions[session.Value])

	expiredCookie := &http.Cookie{
		Name:     "session_id",
		Value:    "",
		HttpOnly: true,
		Secure:   false,
		Expires:  time.Now().AddDate(-1, 0, 0),
		Path:     "/",
	}

	http.SetCookie(w, expiredCookie)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out"})
}
