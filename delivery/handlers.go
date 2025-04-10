package handlery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/usecase"
	"github.com/jackc/pgx/v5"
)

type GetHandler struct {
	DB *pgx.Conn
}

type SessionHandler struct {
	LoginUC        usecase.UserLogIn
	CheckSessionUC usecase.UserCheckSession
	LogoutUC       usecase.UserLogOut
}

type UserHandler struct {
	SignupUC usecase.UserSignUp
}

func CreateCookies(session model.Session) (*model.Cookie, error) {
	cookie := &model.Cookie{
		Name:     "session_id",
		Value:    session.SessionId,
		HttpOnly: true,
		Secure:   false,
		Expires:  time.Now().Add(session.Expires),
		Path:     "/",
	}
	return cookie, nil
}

func (sh *SessionHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid JSON data"})
		return
	}

	if !sh.LoginUC.ValidateLogin(input.Login) || !sh.LoginUC.ValidatePassword(input.Password) {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid login or password"})
		return
	}

	session, err := sh.LoginUC.CreateSession(r.Context(), usecase.LogInInput{
		Login:    input.Login,
		Password: input.Password,
	})
	if err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": fmt.Sprintf("%v", err)})
		return
	}

	cookie, err := CreateCookies(session)
	if err != nil {
		makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "Failed to create cookie"})
		return
	}

	if err := sh.LoginUC.StoreSession(r.Context(), session); err != nil {
		fmt.Println(fmt.Errorf("Error storing session: %v", err))
		makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "Failed to store session"})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookie.Name,
		Value:    cookie.Value,
		HttpOnly: cookie.HttpOnly,
		Secure:   cookie.Secure,
		Expires:  cookie.Expires,
		Path:     cookie.Path,
	})

	makeResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Logged in",
		"user_id": session.UserId,
	})
}

func (uh *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid JSON data"})
		return
	}

	if uh.SignupUC.ValidateLogin(input.Login) != nil || uh.SignupUC.ValidatePassword(input.Password) != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid login or password"})
		return
	}

	if uh.SignupUC.UserExists(r.Context(), input.Login) {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "User already exists"})
		return
	}

	profileId, err := uh.SignupUC.SaveUserProfile(input.Login)
	if err != nil {
		makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "Failed to save user profile"})
		return
	}

	if _, err := uh.SignupUC.SaveUserData(profileId, input.Login, input.Password); err != nil {
		makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "Failed to save user data"})
		return
	}

	makeResponse(w, http.StatusCreated, map[string]string{"message": "User created"})
}

func (sh *SessionHandler) CheckSession(w http.ResponseWriter, r *http.Request) {
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

	userId, err := sh.CheckSessionUC.CheckSession(session.Value)
	if err != nil {
		if err == model.ErrSessionNotFound {
			makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "session not found"})
			return
		}
		if err == model.ErrGetSession {
			makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "error getting session"})
			return
		}
		if err == model.ErrInvalidSessionId {
			makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "error invalid session id"})
			return
		}
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

func (sh *SessionHandler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "No cookies got"})
		return
	}

	if err := sh.LogoutUC.Logout(session.Value); err != nil {
		if err == model.ErrSessionNotFound {
			makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "session not found"})
			return
		}
		if err == model.ErrGetSession {
			makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "error getting session"})
			return
		}
		if err == model.ErrDeleteSession {
			makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "error deleting session"})
			return
		}
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
