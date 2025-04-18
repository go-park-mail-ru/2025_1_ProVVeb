package handlery

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/usecase"
	"github.com/gorilla/mux"
)

type GetHandler struct {
	GetProfileUC    usecase.GetProfile
	GetProfilesUC   usecase.GetProfilesForUser
	GetProfileImage usecase.GetUserPhoto
}

type SessionHandler struct {
	LoginUC        usecase.UserLogIn
	CheckSessionUC usecase.UserCheckSession
	LogoutUC       usecase.UserLogOut
}

type UserHandler struct {
	SignupUC     usecase.UserSignUp
	DeleteUserUC usecase.UserDelete
}

type StaticHandler struct {
	UploadUC usecase.StaticUpload
	DeleteUC usecase.StaticDelete
}

type ProfileHandler struct {
	LikeUC            usecase.ProfileSetLike
	MatchUC           usecase.ProfileGetMatches
	UpdateUC          usecase.ProfileUpdate
	GetProfileUC      usecase.GetProfile
	GetProfileImageUC usecase.GetUserPhoto
}

func (ph *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var profile model.Profile

	// вставить валидацию данных
	// вставит валидацию профиля после middleware

	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid JSON data"})
		return
	}

	profileId := profile.ProfileId

	table_profile, err := ph.GetProfileUC.GetProfile(profileId)
	if err != nil {
		makeResponse(w, http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Error getting profile: %v", err)})
		return
	}

	err = ph.UpdateUC.UpdateProfile(profile, table_profile, profileId)
	if err != nil {
		makeResponse(w, http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Error updating profile: %v", err)})
		return
	}

	makeResponse(w, http.StatusOK, map[string]string{"message": "Updated"})
}

func (ph *ProfileHandler) GetMatches(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	profileId, err := strconv.Atoi(id)
	if err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid user id"})
		return
	}

	profiles, err := ph.MatchUC.GetMatches(profileId)
	if err != nil {
		makeResponse(w, http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Error getting profiles: %v", err)})
		return
	}

	makeResponse(w, http.StatusOK, profiles)
}

func (ph *ProfileHandler) SetLike(w http.ResponseWriter, r *http.Request) {
	var input struct {
		LikeFrom int `json:"likeFrom"`
		LikeTo   int `json:"likeTo"`
		Status   int `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid JSON data"})
		return
	}

	likeFrom, likeTo, status := input.LikeFrom, input.LikeTo, input.Status

	if likeTo == likeFrom {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Please don't like yourself"})
		return
	}

	like_id, err := ph.LikeUC.SetLike(likeFrom, likeTo, status)
	if (like_id == 0) && (err == nil) {
		makeResponse(w, http.StatusConflict, map[string]string{"message": "Already liked"})
		return
	}
	if err != nil {
		makeResponse(w, http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Error getting like: %v", err)})
		return
	}

	makeResponse(w, http.StatusOK, map[string]string{"message": "Liked"})
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

func (sh *StaticHandler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	var maxMemory int64 = model.MaxFileSize
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}

	userId := r.URL.Query().Get("forUser")
	user_id, err := strconv.Atoi(userId)
	if err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": fmt.Sprintf("Invalid user id: %v", err)})
		return
	}

	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": fmt.Sprintf("Invalid multipart form: %v", err)})
		return
	}

	form := r.MultipartForm
	files := form.File["images"]

	if len(files) == 0 {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "No files in 'images' field"})
		return
	}

	var (
		failedUploads  []string
		successUploads []string
	)

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			failedUploads = append(failedUploads, fileHeader.Filename)
			continue
		}
		defer file.Close()

		contentType := fileHeader.Header.Get("Content-Type")
		if !allowedTypes[contentType] {
			failedUploads = append(failedUploads, fileHeader.Filename+" (unsupported type)")
			continue
		}

		buf, err := io.ReadAll(file)
		if err != nil {
			failedUploads = append(failedUploads, fileHeader.Filename+" (read error)")
			continue
		}

		filename := fmt.Sprintf("/%d_%d_%s", user_id, time.Now().UnixNano(), fileHeader.Filename)

		err = sh.UploadUC.UploadUserPhoto(user_id, buf, filename, contentType)
		if err != nil {
			failedUploads = append(failedUploads, fileHeader.Filename+" (upload error)")
			continue
		}

		successUploads = append(successUploads, filename)
	}

	if len(failedUploads) != 0 {
		makeResponse(w, http.StatusInternalServerError, map[string]interface{}{
			"message":        "Some uploads failed",
			"failed_uploads": failedUploads,
		})
		return
	}

	makeResponse(w, http.StatusOK, map[string]interface{}{
		"message":        "All files uploaded",
		"uploaded_files": successUploads,
	})
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

	fmt.Println(fmt.Errorf("%+v", session))

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
		fmt.Println(fmt.Errorf("error storing session: %v", err))
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
		SameSite: http.SameSiteLaxMode,
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
	fmt.Println(fmt.Errorf("cookies^ %+v", session))
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

func (uh *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	userId, err := strconv.Atoi(id)
	if err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid user id"})
		return
	}

	if err := uh.DeleteUserUC.DeleteUser(userId); err != nil {
		makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "Error deleting user"})
		return
	}

	makeResponse(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("User with ID %d deleted", userId)})
}

func (gh *GetHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	profileId, err := strconv.Atoi(id)
	if err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid user id"})
		return
	}

	profile, err := gh.GetProfileUC.GetProfile(profileId)
	if err != nil {
		makeResponse(w, http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Error getting profile: %v", err)})
		return
	}

	makeResponse(w, http.StatusOK, profile)
}

func (gh *GetHandler) GetProfiles(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("forUser")

	profileId, err := strconv.Atoi(userId)
	if err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid user id"})
		return
	}

	profiles, err := gh.GetProfilesUC.GetProfiles(profileId)
	if err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": fmt.Sprintf("Error getting profiles: %v", err)})
		return
	}

	makeResponse(w, http.StatusOK, profiles)
}

func (sh *StaticHandler) DeletePhoto(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	fileURL := r.URL.Query().Get("file_url")

	user_id, err := strconv.Atoi(id)
	if err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid user id"})
		return
	}

	err = sh.DeleteUC.DeleteImage(user_id, fileURL)
	if err != nil {
		makeResponse(w, http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Error deleting photo: %v", err)})
		return
	}

	makeResponse(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("Deleted photo %s for user %d", fileURL, user_id)})

}
