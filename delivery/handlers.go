package handlery

import (
	"encoding/json"
	"net/http"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/usecase"
	"github.com/jackc/pgx/v5"
)

type GetHandler struct {
	DB *pgx.Conn
}

type SessionHandler struct {
	LoginUC usecase.UserLogIn
}

type UserHandler struct {
	DB *pgx.Conn
}

func (u *SessionHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid JSON data"})
		return
	}

	session, err := u.LoginUC.CreateSession(r.Context(), usecase.LogInInput{
		Login:    input.Login,
		Password: input.Password,
	})
	if err != nil {
		handleUseCaseError(w, err)
		return
	}

	cookie, err := u.LoginUC.CreateCookies(r.Context(), session)
	if err != nil {
		makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "Failed to create cookie"})
		return
	}

	if err := u.LoginUC.StoreSession(r.Context(), session); err != nil {
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

// Вспомогательная функция для обработки ошибок use case
func handleUseCaseError(w http.ResponseWriter, err error) {
	switch err {
	case model.ErrInvalidPassword:
		makeResponse(w, http.StatusUnauthorized, map[string]string{"message": "Invalid login or password"})
	// case repository.ErrUserNotFound:
	// 	makeResponse(w, http.StatusNotFound, map[string]string{"message": "User not found"})
	default:
		makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "Login failed"})
	}
}
