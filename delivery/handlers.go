package handlery

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
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
}

type ProfileHandler struct {
	LikeUC          usecase.ProfileSetLike
	MatchUC         usecase.ProfileGetMatches
	GetProfileImage usecase.GetUserPhoto
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
	for i := range profiles {
		photos, err := ph.GetProfileImage.GetUserPhoto(profiles[i].ProfileId)
		if err != nil {
			makeResponse(w, http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Error loading images for profile %d: %v", profiles[i].ProfileId, err)})
			return
		}

		encoded := make([]string, 0, len(photos))
		for _, img := range photos {
			encoded = append(encoded, base64.StdEncoding.EncodeToString(img))
		}
		profiles[i].Photos = encoded
	}

	writer := multipart.NewWriter(w)
	defer writer.Close()

	w.Header().Set("Content-Type", "multipart/form-data; boundary="+writer.Boundary())

	profileJson, err := json.Marshal(profiles)
	if err != nil {
		makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "Error serializing profiles to JSON"})
		return
	}

	part, err := writer.CreateFormField("profiles")
	if err != nil {
		makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "Error creating multipart field"})
		return
	}
	_, err = part.Write(profileJson)
	if err != nil {
		makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "Error writing profile JSON"})
		return
	}

}

func (ph *ProfileHandler) SetLike(w http.ResponseWriter, r *http.Request) {
	var input struct {
		LikeFrom string `json:"login"`
		LikedBy  string `json:"password"`
		Status   string `json:"Status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid JSON data"})
		return
	}

	LikeFrom, err := strconv.Atoi(input.LikeFrom)
	if err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid user id"})
		return
	}

	Status, err := strconv.Atoi(input.Status)
	if (err != nil) || ((Status != 1) && (Status != -1)) {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid status"})
		return
	}

	LikeBy, err := strconv.Atoi(input.LikeFrom)
	if err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid user id"})
		return
	}

	if LikeBy == LikeFrom {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Please dont like yourself"})
		return
	}

	err = ph.LikeUC.SetLike(LikeBy, LikeFrom, Status)
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
	const maxMemory = 10 << 20
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}

	userId := r.URL.Query().Get("forUser")
	user_id, err := strconv.Atoi(userId)
	if err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid user id"})
		return
	}

	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid multipart form"})
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

		filename := fmt.Sprintf("%d_%d_%s", user_id, time.Now().UnixNano(), fileHeader.Filename)

		err = sh.UploadUC.UploadUserPhoto(user_id, buf, filename, contentType)
		if err != nil {
			failedUploads = append(failedUploads, fileHeader.Filename+" (upload error)")
			continue
		}

		successUploads = append(successUploads, filename)
	}

	if len(failedUploads) == 0 {
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

// первый раз пришли на сервис
// делаем запрос checkSession
// сессии нет, поэтому логинимся
// выполняется логин, успех
// пользователь что-то делает - sessionId хранится в куках
// как только страница обновляется - куки становятся просроченнные ИЛИ...

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

	files, err := gh.GetProfileImage.GetUserPhoto(profileId)
	if err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": fmt.Sprintf("error loading images: %v", err)})
		return
	}

	writer := multipart.NewWriter(w)
	defer writer.Close()

	w.Header().Set("Content-Type", "multipart/form-data; boundary="+writer.Boundary())

	jsonData, err := json.Marshal(profile)
	if err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Failed to marshal profile"})
		return
	}

	jsonPart, err := writer.CreateFormField("profile")
	if err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Failed to create profile part"})
		return
	}
	_, err = jsonPart.Write(jsonData)
	if err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Failed to write profile part"})
		return
	}

	for i, file := range files {
		part, err := writer.CreateFormFile(fmt.Sprintf("photo%d", i+1), fmt.Sprintf("photo%d.jpg", i+1))
		if err != nil {
			makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Failed to create image part"})
			return
		}
		_, err = part.Write(file)
		if err != nil {
			makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Failed to write image data"})
			return
		}
	}
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
	for i := range profiles {
		photos, err := gh.GetProfileImage.GetUserPhoto(profiles[i].ProfileId)
		if err != nil {
			makeResponse(w, http.StatusBadRequest, map[string]string{"message": fmt.Sprintf("Error loading images for profile %d: %v", profiles[i].ProfileId, err)})
			return
		}

		encoded := make([]string, 0, len(photos))
		for _, img := range photos {
			encoded = append(encoded, base64.StdEncoding.EncodeToString(img))
		}
		profiles[i].Photos = encoded
	}

	writer := multipart.NewWriter(w)
	defer writer.Close()

	w.Header().Set("Content-Type", "multipart/form-data; boundary="+writer.Boundary())

	profileJson, err := json.Marshal(profiles)
	if err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Error serializing profiles to JSON"})
		return
	}

	part, err := writer.CreateFormField("profiles")
	if err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Error creating multipart field"})
		return
	}
	_, err = part.Write(profileJson)
	if err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Error writing profile JSON"})
		return
	}

}
