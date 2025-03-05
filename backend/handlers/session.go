package handlers

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/backend/utils"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/config"
)

var Se = struct {
	users map[int]config.User
}{users: utils.InitUserMap()}

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

	login := r.FormValue("login")
	password := r.FormValue("password")

	userId, err := strconv.Atoi(login)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, ok := Se.users[userId]
	if !ok {
		http.Error(w, "No such user", http.StatusNotFound)
		return
	}

	if user.Password != password {
		http.Error(w, "Invalid password", http.StatusBadRequest)
		return
	}

	SID := RandStringRunes(32)

	api.sessions[SID] = user.Id

	cookie := &http.Cookie{
		Name:    "session_id",
		Value:   SID,
		Expires: time.Now().Add(10 * time.Hour),
	}
	http.SetCookie(w, cookie)

	w.Write([]byte(SID))
}

func (u *SessionHandler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if _, ok := api.sessions[session.Value]; !ok {
		http.Error(w, "Session not found", http.StatusUnauthorized)
		return
	}

	delete(api.sessions, session.Value)

	session.Expires = time.Now().AddDate(0, 0, -1)
	http.SetCookie(w, session)

	http.Redirect(w, r, "/", http.StatusFound)
}
