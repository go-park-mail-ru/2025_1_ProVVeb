package handlers

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/backend/utils"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/config"
)

var Se = struct {
	users map[int]config.User
}{users: utils.InitUserMap()}

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
		Email    string
		Password string
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&gotData); err != nil {
		http.Error(w, "Invalid login or password", http.StatusBadRequest)
	}

	emal, password := gotData.Email, gotData.Password

	var foundUser *config.User
	for _, user := range Se.users {
		if user.Email == emal {
			foundUser = &user
			break
		}
	}
	if foundUser == nil {
		http.Error(w, "No such user", http.StatusBadRequest)
		return
	}

	if foundUser.Password != utils.EncryptPasswordSHA256(password) {
		http.Error(w, "Invalid password", http.StatusBadRequest)
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
	}

	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged in"})
}

func (u *SessionHandler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		http.Error(w, "No cookies got", http.StatusBadRequest)
		return
	}

	if _, ok := api.sessions[session.Value]; !ok {
		http.Error(w, "Session not found", http.StatusUnauthorized)
		return
	}

	delete(api.sessions, session.Value)

	session.Expires = time.Now().AddDate(-1, 0, 0)
	http.SetCookie(w, session)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out"})
}
