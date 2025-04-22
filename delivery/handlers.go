package handlers

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
	"github.com/microcosm-cc/bluemonday"
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
	DeleteUC usecase.DeleteStatic
}

type ProfileHandler struct {
	LikeUC            usecase.ProfileSetLike
	MatchUC           usecase.GetProfileMatches
	UpdateUC          usecase.ProfileUpdate
	GetProfileUC      usecase.GetProfile
	GetProfileImageUC usecase.GetUserPhoto
}

func (ph *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userIDRaw := r.Context().Value(userIDKey)
	profileId, ok := userIDRaw.(int)
	if !ok {
		MakeResponse(w, http.StatusUnauthorized, map[string]string{"message": "You don't have access"})
		return
	}

	var profile model.Profile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		MakeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid JSON data"})
		return
	}
	if profileId != profile.ProfileId {
		MakeResponse(w, http.StatusUnauthorized, map[string]string{"message": "You don't have access for this"})
		return
	}

	table_profile, err := ph.GetProfileUC.GetProfile(profileId)
	if err != nil {
		MakeResponse(w, http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Error getting profile: %v", err)})
		return
	}

	err = ph.UpdateUC.UpdateProfile(profile, table_profile, profileId)
	if err != nil {
		MakeResponse(w, http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Error updating profile: %v", err)})
		return
	}

	MakeResponse(w, http.StatusOK, map[string]string{"message": "Updated"})
}

func (ph *ProfileHandler) GetMatches(w http.ResponseWriter, r *http.Request) {
	userIDRaw := r.Context().Value(userIDKey)
	profileId, ok := userIDRaw.(int)
	if !ok {
		MakeResponse(w, http.StatusUnauthorized, map[string]string{"message": "You don't have access"})
		return
	}

	profiles, err := ph.MatchUC.GetMatches(profileId)
	if err != nil {
		MakeResponse(w, http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Error getting profiles: %v", err)})
		return
	}

	MakeResponse(w, http.StatusOK, profiles)
}

func (ph *ProfileHandler) SetLike(w http.ResponseWriter, r *http.Request) {
	userIDRaw := r.Context().Value(userIDKey)
	profileId, ok := userIDRaw.(int)
	if !ok {
		MakeResponse(w, http.StatusUnauthorized, map[string]string{"message": "You don't have access"})
		return
	}

	var input struct {
		LikeFrom int `json:"likeFrom"`
		LikeTo   int `json:"likeTo"`
		Status   int `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		MakeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid JSON data"})
		return
	}

	likeFrom := input.LikeFrom
	likeTo := input.LikeTo
	status := input.Status

	if likeTo == likeFrom {
		MakeResponse(w, http.StatusBadRequest, map[string]string{"message": "Please don't like yourself"})
		return
	}

	if profileId != likeFrom {
		MakeResponse(w, http.StatusBadRequest, map[string]string{"message": "You are unauthorized to like this user"})
		return
	}

	like_id, err := ph.LikeUC.SetLike(likeFrom, likeTo, status)
	if (like_id == 0) && (err == nil) {
		MakeResponse(w, http.StatusConflict, map[string]string{"message": "Already liked"})
		return
	}
	if err != nil {
		MakeResponse(w, http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Error getting like: %v", err)})
		return
	}

	MakeResponse(w, http.StatusOK, map[string]string{"message": "Liked"})
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
	sanitizer := bluemonday.UGCPolicy()
	var maxMemory int64 = model.MaxFileSize
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}

	userIDRaw := r.Context().Value(userIDKey)
	user_id, ok := userIDRaw.(int)
	if !ok {
		MakeResponse(w, http.StatusUnauthorized, map[string]string{"message": "You don't have access"})
		return
	}

	err := r.ParseMultipartForm(maxMemory)
	if err != nil {
		MakeResponse(w, http.StatusBadRequest, map[string]string{"message": fmt.Sprintf("Invalid multipart form: %v", err)})
		return
	}

	form := r.MultipartForm
	files := form.File["images"]

	if len(files) == 0 {
		MakeResponse(w, http.StatusBadRequest, map[string]string{"message": "No files in 'images' field"})
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

		contentType := sanitizer.Sanitize(fileHeader.Header.Get("Content-Type"))
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
		MakeResponse(w, http.StatusInternalServerError, map[string]interface{}{
			"message":        "Some uploads failed",
			"failed_uploads": failedUploads,
		})
		return
	}

	MakeResponse(w, http.StatusOK, map[string]interface{}{
		"message":        "All files uploaded",
		"uploaded_files": successUploads,
	})
}

func (sh *SessionHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	sanitizer := bluemonday.UGCPolicy()
	var input struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		MakeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid JSON data"})
		return
	}

	input.Login = sanitizer.Sanitize(input.Login)
	input.Password = sanitizer.Sanitize(input.Password)

	if !sh.LoginUC.ValidateLogin(input.Login) || !sh.LoginUC.ValidatePassword(input.Password) {
		MakeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid login or password"})
		return
	}

	session, err := sh.LoginUC.CreateSession(r.Context(), usecase.LogInInput{
		Login:    input.Login,
		Password: input.Password,
	})

	// fmt.Println(fmt.Errorf("%+v", session))

	if err != nil {
		MakeResponse(w, http.StatusBadRequest, map[string]string{"message": fmt.Sprintf("%v", err)})
		return
	}

	cookie, err := CreateCookies(session)
	if err != nil {
		MakeResponse(w, http.StatusInternalServerError, map[string]string{"message": "Failed to create cookie"})
		return
	}

	if err := sh.LoginUC.StoreSession(r.Context(), session); err != nil {
		fmt.Println(fmt.Errorf("error storing session: %v", err))
		MakeResponse(w, http.StatusInternalServerError, map[string]string{"message": "Failed to store session"})
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

	MakeResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Logged in",
		"user_id": session.UserId,
	})
}

func (uh *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	sanitizer := bluemonday.UGCPolicy()
	var input struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		MakeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid JSON data"})
		return
	}

	input.Login = sanitizer.Sanitize(input.Login)
	input.Password = sanitizer.Sanitize(input.Password)

	if uh.SignupUC.ValidateLogin(input.Login) != nil || uh.SignupUC.ValidatePassword(input.Password) != nil {
		MakeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid login or password"})
		return
	}

	if uh.SignupUC.UserExists(r.Context(), input.Login) {
		MakeResponse(w, http.StatusBadRequest, map[string]string{"message": "User already exists"})
		return
	}

	profileId, err := uh.SignupUC.SaveUserProfile(input.Login)
	if err != nil {
		MakeResponse(w, http.StatusInternalServerError, map[string]string{"message": "Failed to save user profile"})
		return
	}

	if _, err := uh.SignupUC.SaveUserData(profileId, input.Login, input.Password); err != nil {
		MakeResponse(w, http.StatusInternalServerError, map[string]string{"message": "Failed to save user data"})
		return
	}

	MakeResponse(w, http.StatusCreated, map[string]string{"message": "User created"})
}

func (sh *SessionHandler) CheckSession(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session_id")
	// fmt.Println(fmt.Errorf("cookies^ %+v", session))
	if err == http.ErrNoCookie {
		response := struct {
			Message   string `json:"message"`
			InSession bool   `json:"inSession"`
		}{
			Message:   "No cookies got",
			InSession: false,
		}
		MakeResponse(w, http.StatusOK, response)
		return
	}

	userId, err := sh.CheckSessionUC.CheckSession(session.Value)
	if err != nil {
		if err == model.ErrSessionNotFound {
			MakeResponse(w, http.StatusInternalServerError, map[string]string{"message": "session not found"})
			return
		}
		if err == model.ErrGetSession {
			MakeResponse(w, http.StatusInternalServerError, map[string]string{"message": "error getting session"})
			return
		}
		if err == model.ErrInvalidSessionId {
			MakeResponse(w, http.StatusInternalServerError, map[string]string{"message": "error invalid session id"})
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

	MakeResponse(w, http.StatusOK, response)
}

func (sh *SessionHandler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		MakeResponse(w, http.StatusBadRequest, map[string]string{"message": "No cookies got"})
		return
	}

	if err := sh.LogoutUC.Logout(session.Value); err != nil {
		if err == model.ErrSessionNotFound {
			MakeResponse(w, http.StatusInternalServerError, map[string]string{"message": "session not found"})
			return
		}
		if err == model.ErrGetSession {
			MakeResponse(w, http.StatusInternalServerError, map[string]string{"message": "error getting session"})
			return
		}
		if err == model.ErrDeleteSession {
			MakeResponse(w, http.StatusInternalServerError, map[string]string{"message": "error deleting session"})
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

	MakeResponse(w, http.StatusOK, map[string]string{"message": "Logged out"})
}

func (uh *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	sanitizer := bluemonday.UGCPolicy()
	vars := mux.Vars(r)
	id := vars["id"]

	userId, err := strconv.Atoi(sanitizer.Sanitize(id))
	if err != nil {
		MakeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid user id"})
		return
	}

	if err := uh.DeleteUserUC.DeleteUser(userId); err != nil {
		MakeResponse(w, http.StatusInternalServerError, map[string]string{"message": "Error deleting user"})
		return
	}

	MakeResponse(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("User with ID %d deleted", userId)})
}

func (gh *GetHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userIDRaw := r.Context().Value(userIDKey)
	fmt.Println(userIDRaw)
	profileId, ok := userIDRaw.(uint32)
	if !ok {
		MakeResponse(w, http.StatusUnauthorized, map[string]string{"message": "You don't have access"})
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fmt.Println(userIDRaw)

	profile, err := gh.GetProfileUC.GetProfile(int(profileId))
	if err != nil {
		MakeResponse(w, http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Error getting profile: %v", err)})
		return
	}

	MakeResponse(w, http.StatusOK, profile)
}

func (gh *GetHandler) GetProfiles(w http.ResponseWriter, r *http.Request) {
	userIDRaw := r.Context().Value(userIDKey)
	profileId, ok := userIDRaw.(int)
	if !ok {
		MakeResponse(w, http.StatusUnauthorized, map[string]string{"message": "You don't have access"})
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profiles, err := gh.GetProfilesUC.GetProfiles(profileId)
	if err != nil {
		MakeResponse(w, http.StatusBadRequest, map[string]string{"message": fmt.Sprintf("Error getting profiles: %v", err)})
		return
	}

	MakeResponse(w, http.StatusOK, profiles)
}

func (sh *StaticHandler) DeletePhoto(w http.ResponseWriter, r *http.Request) {
	sanitizer := bluemonday.UGCPolicy()
	fileURL := sanitizer.Sanitize(r.URL.Query().Get("file_url"))

	userIDRaw := r.Context().Value(userIDKey)
	user_id, ok := userIDRaw.(int)
	if !ok {
		MakeResponse(w, http.StatusUnauthorized, map[string]string{"message": "You don't have access"})
		return
	}

	err := sh.DeleteUC.DeleteImage(user_id, fileURL)
	if err != nil {
		MakeResponse(w, http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Error deleting photo: %v", err)})
		return
	}

	MakeResponse(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("Deleted photo %s for user %d", fileURL, user_id)})

}
