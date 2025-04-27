package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/usecase"

	"github.com/gorilla/mux"
	"github.com/microcosm-cc/bluemonday"
	"github.com/sirupsen/logrus"
)

type GetHandler struct {
	GetProfileUC    usecase.GetProfile
	GetProfilesUC   usecase.GetProfilesForUser
	GetProfileImage usecase.GetUserPhoto
	Logger          *logger.LogrusLogger
}

type SessionHandler struct {
	LoginUC        usecase.UserLogIn
	CheckSessionUC usecase.UserCheckSession
	LogoutUC       usecase.UserLogOut
	Logger         *logger.LogrusLogger
}

type UserHandler struct {
	SignupUC     usecase.UserSignUp
	DeleteUserUC usecase.UserDelete
	Logger       *logger.LogrusLogger
}

type StaticHandler struct {
	UploadUC usecase.StaticUpload
	DeleteUC usecase.DeleteStatic
	Logger   *logger.LogrusLogger
}

type ProfileHandler struct {
	LikeUC            usecase.ProfileSetLike
	MatchUC           usecase.GetProfileMatches
	UpdateUC          usecase.ProfileUpdate
	GetProfileUC      usecase.GetProfile
	GetProfileImageUC usecase.GetUserPhoto
	Logger            *logger.LogrusLogger
}

type QueryHandler struct {
	GetActiveQueriesUC   usecase.GetActiveQueries
	StoreUserAnswerUC    usecase.StoreUserAnswer
	GetAnswersForUserUC  usecase.GetAnswersForUser
	GetAnswersForQueryUC usecase.GetAnswersForQuery
	Logger               *logger.LogrusLogger
}

func (ph *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ph.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
	}).Info("start processing UpdateProfile request")

	userIDRaw := r.Context().Value(userIDKey)
	profileId, ok := userIDRaw.(uint32)
	if !ok {
		ph.Logger.WithFields(&logrus.Fields{
			"error": "failed to get userID from context",
		}).Warn("unauthorized access attempt")

		MakeResponse(w, http.StatusUnauthorized,
			map[string]string{"message": "You don't have access"},
		)
		return
	}

	var profile model.Profile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"error":      err.Error(),
			"profile_id": profileId,
		}).Warn("failed to decode request body")

		MakeResponse(w, http.StatusBadRequest,
			map[string]string{"message": "Invalid JSON data"},
		)
		return
	}

	if profile.ProfileId != 0 && int(profileId) != profile.ProfileId {
		ph.Logger.WithFields(&logrus.Fields{
			"request_profile_id":  profileId,
			"provided_profile_id": profile.ProfileId,
		}).Warn("profile ID mismatch")

		MakeResponse(w, http.StatusUnauthorized,
			map[string]string{"message": "You don't have access for this"},
		)
		return
	}

	table_profile, err := ph.GetProfileUC.GetProfile(int(profileId))
	if err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"error":      err.Error(),
		}).Error("failed to get profile from database")

		MakeResponse(w, http.StatusInternalServerError,
			map[string]string{"message": fmt.Sprintf("Error getting profile: %v", err)},
		)
		return
	}

	err = ph.UpdateUC.UpdateProfile(profile, table_profile, int(profileId))
	if err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"profile_id":   profileId,
			"error":        err.Error(),
			"profile_data": profile,
		}).Error("failed to update profile")

		MakeResponse(w, http.StatusInternalServerError,
			map[string]string{"message": fmt.Sprintf("Error updating profile: %v", err)},
		)
		return
	}

	ph.Logger.WithFields(&logrus.Fields{
		"profile_id": profileId,
	}).Info("profile updated successfully")

	MakeResponse(w, http.StatusOK, map[string]string{"message": "Updated"})
}

func (ph *ProfileHandler) GetMatches(w http.ResponseWriter, r *http.Request) {
	ph.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
	}).Info("GetMatches request started")

	userIDRaw := r.Context().Value(userIDKey)
	profileId, ok := userIDRaw.(uint32)
	if !ok {
		ph.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized access attempt")

		MakeResponse(w, http.StatusUnauthorized,
			map[string]string{"message": "You don't have access"},
		)
		return
	}

	ph.Logger.WithFields(&logrus.Fields{
		"profile_id": profileId,
	}).Debug("attempting to get matches")

	profiles, err := ph.MatchUC.GetMatches(int(profileId))
	if err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"error":      err.Error(),
		}).Error("failed to get matches")

		MakeResponse(w, http.StatusInternalServerError,
			map[string]string{"message": fmt.Sprintf("Error getting profiles: %v", err)},
		)
		return
	}

	ph.Logger.WithFields(&logrus.Fields{
		"profile_id":    profileId,
		"matches_count": len(profiles),
	}).Info("successfully retrieved matches")

	MakeResponse(w, http.StatusOK, profiles)
}

func (ph *ProfileHandler) SetLike(w http.ResponseWriter, r *http.Request) {
	ph.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
	}).Info("SetLike request started")

	userIDRaw := r.Context().Value(userIDKey)
	profileId, ok := userIDRaw.(uint32)
	if !ok {
		ph.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized access attempt")

		MakeResponse(w, http.StatusUnauthorized,
			map[string]string{"message": "You don't have access"},
		)
		return
	}

	var input struct {
		LikeFrom int `json:"likeFrom"`
		LikeTo   int `json:"likeTo"`
		Status   int `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"error":      err.Error(),
		}).Warn("failed to decode like request body")

		MakeResponse(w, http.StatusBadRequest,
			map[string]string{"message": "Invalid JSON data"},
		)
		return
	}

	likeFrom := input.LikeFrom
	likeTo := input.LikeTo
	status := input.Status

	ph.Logger.WithFields(&logrus.Fields{
		"profile_id": profileId,
		"like_from":  likeFrom,
		"like_to":    likeTo,
		"status":     status,
	}).Debug("processing like action")

	if likeTo == likeFrom {
		ph.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
		}).Warn("attempt to like oneself")

		MakeResponse(w, http.StatusBadRequest, map[string]string{"message": "Please don't like yourself"})
		return
	}

	if int(profileId) != likeFrom {
		ph.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"like_from":  likeFrom,
		}).Warn("unauthorized like attempt")

		MakeResponse(w, http.StatusBadRequest, map[string]string{"message": "You are unauthorized to like this user"})
		return
	}

	like_id, err := ph.LikeUC.SetLike(likeFrom, likeTo, status)
	if (like_id == 0) && (err == nil) {
		ph.Logger.WithFields(&logrus.Fields{
			"like_from": likeFrom,
			"like_to":   likeTo,
		}).Info("duplicate like detected")

		MakeResponse(w, http.StatusConflict, map[string]string{"message": "Already liked"})
		return
	}
	if err != nil {
		ph.Logger.WithFields(&logrus.Fields{
			"like_from": likeFrom,
			"like_to":   likeTo,
			"error":     err.Error(),
		}).Error("failed to set like")

		MakeResponse(w, http.StatusInternalServerError, map[string]string{"message": fmt.Sprintf("Error getting like: %v", err)})
		return
	}

	ph.Logger.WithFields(&logrus.Fields{
		"like_id":   like_id,
		"like_from": likeFrom,
		"like_to":   likeTo,
		"status":    status,
	}).Info("like successfully processed")

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
	sh.Logger.WithFields(&logrus.Fields{
		"method":       r.Method,
		"path":         r.URL.Path,
		"request_id":   r.Header.Get("request_id"),
		"content_type": r.Header.Get("Content-Type"),
	}).Info("UploadPhoto request started")

	sanitizer := bluemonday.UGCPolicy()
	var maxMemory int64 = model.MaxFileSize
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}

	userIDRaw := r.Context().Value(userIDKey)
	user_id, ok := userIDRaw.(uint32)
	if !ok {
		sh.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized upload attempt")

		MakeResponse(w, http.StatusUnauthorized,
			map[string]string{"message": "You don't have access"},
		)
		return
	}

	sh.Logger.WithFields(&logrus.Fields{
		"user_id":    user_id,
		"max_memory": maxMemory,
	}).Debug("parsing multipart form")

	err := r.ParseMultipartForm(maxMemory)
	if err != nil {
		sh.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Warn("failed to parse multipart form")

		MakeResponse(w, http.StatusBadRequest,
			map[string]string{"message": fmt.Sprintf("Invalid multipart form: %v", err)},
		)
		return
	}

	form := r.MultipartForm
	files := form.File["images"]

	if len(files) == 0 {
		sh.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
		}).Warn("no files in 'images' field")

		MakeResponse(w, http.StatusBadRequest,
			map[string]string{"message": "No files in 'images' field"},
		)
		return
	}

	sh.Logger.WithFields(&logrus.Fields{
		"user_id":       user_id,
		"files_count":   len(files),
		"allowed_types": allowedTypes,
	}).Info("starting files processing")

	var (
		failedUploads  []string
		successUploads []string
	)

	for _, fileHeader := range files {
		fileName := fileHeader.Filename
		fileSize := fileHeader.Size
		contentType := fileHeader.Header.Get("Content-Type")

		sh.Logger.WithFields(&logrus.Fields{
			"file_name":    fileName,
			"file_size":    fileSize,
			"content_type": contentType,
		}).Debug("processing file")

		file, err := fileHeader.Open()
		if err != nil {
			sh.Logger.WithFields(&logrus.Fields{
				"file_name": fileName,
				"error":     err.Error(),
			}).Warn("failed to open file")

			failedUploads = append(failedUploads, fileName)
			continue
		}
		defer file.Close()

		sanitizedType := sanitizer.Sanitize(contentType)
		if !allowedTypes[sanitizedType] {
			sh.Logger.WithFields(&logrus.Fields{
				"file_name":    fileName,
				"content_type": sanitizedType,
			}).Warn("unsupported file type")

			failedUploads = append(failedUploads, fileName+" (unsupported type)")
			continue
		}

		buf, err := io.ReadAll(file)
		if err != nil {
			sh.Logger.WithFields(&logrus.Fields{
				"file_name": fileName,
				"error":     err.Error(),
			}).Warn("failed to read file content")

			failedUploads = append(failedUploads, fileName+" (read error)")
			continue
		}

		filename := fmt.Sprintf("/%d_%d_%s", user_id, time.Now().UnixNano(), fileName)

		sh.Logger.WithFields(&logrus.Fields{
			"user_id":   user_id,
			"file_name": filename,
			"data_size": len(buf),
		}).Debug("uploading file to storage")

		err = sh.UploadUC.UploadUserPhoto(int(user_id), buf, filename, sanitizedType)
		if err != nil {
			sh.Logger.WithFields(&logrus.Fields{
				"user_id":   user_id,
				"file_name": filename,
				"error":     err.Error(),
			}).Error("failed to upload file")

			failedUploads = append(failedUploads, fileName+" (upload error)")
			continue
		}

		successUploads = append(successUploads, filename)
		sh.Logger.WithFields(&logrus.Fields{
			"user_id":   user_id,
			"file_name": filename,
		}).Info("file uploaded successfully")
	}

	sh.Logger.WithFields(&logrus.Fields{
		"user_id":        user_id,
		"total_files":    len(files),
		"success_count":  len(successUploads),
		"failed_count":   len(failedUploads),
		"failed_uploads": failedUploads,
	}).Info("files processing completed")

	if len(failedUploads) != 0 {
		if len(successUploads) > 0 {
			sh.Logger.WithFields(&logrus.Fields{
				"user_id":       user_id,
				"success_count": len(successUploads),
				"failed_count":  len(failedUploads),
			}).Warn("partial upload failure")
		} else {
			sh.Logger.WithFields(&logrus.Fields{
				"user_id":      user_id,
				"failed_count": len(failedUploads),
			}).Error("all files failed to upload")
		}

		MakeResponse(w, http.StatusInternalServerError, map[string]interface{}{
			"message":        "Some uploads failed",
			"failed_uploads": failedUploads,
		})
		return
	}

	sh.Logger.WithFields(&logrus.Fields{
		"user_id":        user_id,
		"uploaded_files": successUploads,
	}).Info("all files uploaded successfully")

	MakeResponse(w, http.StatusOK, map[string]interface{}{
		"message":        "All files uploaded",
		"uploaded_files": successUploads,
	})
}

func (sh *SessionHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	sh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("Login request started")

	sanitizer := bluemonday.UGCPolicy()
	var input struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		sh.Logger.WithFields(&logrus.Fields{
			"error": err.Error(),
		}).Warn("failed to decode login request body")

		MakeResponse(w, http.StatusBadRequest,
			map[string]string{"message": "Invalid JSON data"},
		)
		return
	}

	input.Login = sanitizer.Sanitize(input.Login)
	input.Password = sanitizer.Sanitize(input.Password)

	if !sh.LoginUC.ValidateLogin(input.Login) || !sh.LoginUC.ValidatePassword(input.Password) {
		sh.Logger.WithFields(&logrus.Fields{
			"login": input.Login,
			"error": "invalid login or password format",
		}).Warn("validation failed")

		MakeResponse(w, http.StatusBadRequest,
			map[string]string{"message": "Invalid login or password"},
		)
		return
	}

	ip := r.RemoteAddr
	err := sh.LoginUC.CheckAttempts(r.Context(), ip)
	if err != nil {
		sh.Logger.WithFields(&logrus.Fields{
			"ip":    ip,
			"error": "too many login attempts",
		}).Warn("login attempts limit exceeded")

		MakeResponse(w, http.StatusBadRequest,
			map[string]string{"message": "Дядь, хватит дудосить, ты забыл пароль"},
		)
		return
	}

	sh.Logger.WithFields(&logrus.Fields{
		"login": input.Login,
	}).Debug("attempting to create session")

	session, err := sh.LoginUC.CreateSession(r.Context(), usecase.LogInInput{
		Login:    input.Login,
		Password: input.Password,
	})
	if err != nil {
		sh.Logger.WithFields(&logrus.Fields{
			"login": input.Login,
			"ip":    ip,
			"error": err.Error(),
		}).Warn("failed to create session")

		sh.LoginUC.IncreaseAttempts(r.Context(), ip)
		MakeResponse(w, http.StatusBadRequest,
			map[string]string{"message": fmt.Sprintf("%v", err)},
		)
		return
	}

	sh.Logger.WithFields(&logrus.Fields{
		"user_id":    session.UserId,
		"session_id": session.SessionId,
	}).Info("user authenticated successfully")

	cookie, err := CreateCookies(session)
	if err != nil {
		sh.Logger.WithFields(&logrus.Fields{
			"user_id": session.UserId,
			"error":   err.Error(),
		}).Error("failed to create session cookie")

		MakeResponse(w, http.StatusInternalServerError,
			map[string]string{"message": "Failed to create cookie"},
		)
		return
	}

	if err := sh.LoginUC.StoreSession(r.Context(), session); err != nil {
		sh.Logger.WithFields(&logrus.Fields{
			"session_id": session.SessionId,
			"user_id":    session.UserId,
			"error":      err.Error(),
		}).Error("failed to store session")

		MakeResponse(w, http.StatusInternalServerError,
			map[string]string{"message": "Failed to store session"},
		)
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

	token, _ := sh.LoginUC.CreateJwtToken(&repository.Session{
		ID:     session.SessionId,
		UserID: uint32(session.UserId),
	}, time.Now().Add(12*time.Hour).Unix())

	sh.Logger.WithFields(&logrus.Fields{
		"user_id": session.UserId,
	}).Debug("JWT token created")

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    token,
		HttpOnly: false,
		Secure:   false,
		Path:     "/",
		Expires:  time.Now().Add(12 * time.Hour),
		SameSite: http.SameSiteLaxMode,
	})

	_ = sh.LoginUC.DeleteAttempts(r.Context(), ip)

	sh.Logger.WithFields(&logrus.Fields{
		"user_id":    session.UserId,
		"session_id": session.SessionId,
		"ip":         ip,
	}).Info("login completed successfully")

	MakeResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Logged in",
		"user_id": session.UserId,
	})
}

func (uh *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	uh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("CreateUser request started")

	sanitizer := bluemonday.UGCPolicy()
	var input struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		uh.Logger.WithFields(&logrus.Fields{
			"error": err.Error(),
		}).Warn("failed to decode user creation request")

		MakeResponse(w, http.StatusBadRequest,
			map[string]string{"message": "Invalid JSON data"},
		)
		return
	}

	input.Login = sanitizer.Sanitize(input.Login)
	input.Password = sanitizer.Sanitize(input.Password)

	if err := uh.SignupUC.ValidateLogin(input.Login); err != nil {
		uh.Logger.WithFields(&logrus.Fields{
			"login": input.Login,
			"error": err.Error(),
		}).Warn("login validation failed")

		MakeResponse(w, http.StatusBadRequest,
			map[string]string{"message": "Invalid login or password"},
		)
		return
	}

	if err := uh.SignupUC.ValidatePassword(input.Password); err != nil {
		uh.Logger.WithFields(&logrus.Fields{
			"login": input.Login,
			"error": err.Error(),
		}).Warn("password validation failed")

		MakeResponse(w, http.StatusBadRequest,
			map[string]string{"message": "Invalid login or password"},
		)
		return
	}

	if uh.SignupUC.UserExists(r.Context(), input.Login) {
		uh.Logger.WithFields(&logrus.Fields{
			"login": input.Login,
		}).Warn("user already exists")

		MakeResponse(w, http.StatusBadRequest,
			map[string]string{"message": "User already exists"},
		)
		return
	}

	profileId, err := uh.SignupUC.SaveUserProfile(input.Login)
	if err != nil {
		uh.Logger.WithFields(&logrus.Fields{
			"login": input.Login,
			"error": err.Error(),
		}).Error("failed to save user profile")

		MakeResponse(w, http.StatusInternalServerError,
			map[string]string{"message": "Failed to save user profile"},
		)
		return
	}

	uh.Logger.WithFields(&logrus.Fields{
		"profile_id": profileId,
		"login":      input.Login,
	}).Debug("user profile created")

	if _, err := uh.SignupUC.SaveUserData(profileId, input.Login, input.Password); err != nil {
		uh.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"login":      input.Login,
			"error":      err.Error(),
		}).Error("failed to save user data")

		MakeResponse(w, http.StatusInternalServerError,
			map[string]string{"message": "Failed to save user data"},
		)
		return
	}

	uh.Logger.WithFields(&logrus.Fields{
		"profile_id": profileId,
		"login":      input.Login,
	}).Info("user created successfully")

	MakeResponse(w, http.StatusCreated,
		map[string]string{"message": "User created"},
	)
}

func (sh *SessionHandler) CheckSession(w http.ResponseWriter, r *http.Request) {
	sh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("CheckSession request started")

	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		sh.Logger.WithFields(&logrus.Fields{
			"error": "no session cookie found",
		}).Debug("session check failed - no cookies")

		response := struct {
			Message   string `json:"message"`
			InSession bool   `json:"inSession"`
		}{
			Message:   "No cookies got",
			InSession: false,
		}
		MakeResponse(w, http.StatusOK, response)
		return
	} else if err != nil {
		sh.Logger.WithFields(&logrus.Fields{
			"error": err.Error(),
		}).Warn("failed to get session cookie")

		MakeResponse(w, http.StatusBadRequest,
			map[string]string{"message": "Invalid cookie"},
		)
		return
	}

	userId, err := sh.CheckSessionUC.CheckSession(session.Value)
	if err != nil {
		if err == model.ErrSessionNotFound {
			sh.Logger.WithFields(&logrus.Fields{
				"session_id": session.Value,
				"error":      err.Error(),
			}).Warn("session not found")

			MakeResponse(w, http.StatusInternalServerError,
				map[string]string{"message": "session not found"},
			)
			return
		}
		if err == model.ErrGetSession {
			sh.Logger.WithFields(&logrus.Fields{
				"session_id": session.Value,
				"error":      err.Error(),
			}).Error("failed to get session")

			MakeResponse(w, http.StatusInternalServerError,
				map[string]string{"message": "error getting session"},
			)
			return
		}
		if err == model.ErrInvalidSessionId {
			sh.Logger.WithFields(&logrus.Fields{
				"session_id": session.Value,
				"error":      err.Error(),
			}).Warn("invalid session id")
			MakeResponse(w, http.StatusInternalServerError,
				map[string]string{"message": "error invalid session id"},
			)
			return
		}

		sh.Logger.WithFields(&logrus.Fields{
			"session_id": session.Value,
			"error":      err.Error(),
		}).Error("unknown session check error")

		MakeResponse(w, http.StatusInternalServerError,
			map[string]string{"message": "unknown session error"},
		)
		return
	}

	sh.Logger.WithFields(&logrus.Fields{
		"user_id":    userId,
		"session_id": session.Value,
	}).Info("session check successful")

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
	sh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("Logout request started")

	session, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		sh.Logger.WithFields(&logrus.Fields{
			"error": "session cookie not found",
		}).Warn("logout attempt without session cookie")

		MakeResponse(w, http.StatusBadRequest,
			map[string]string{"message": "No cookies got"},
		)
		return
	} else if err != nil {
		sh.Logger.WithFields(&logrus.Fields{
			"error": err.Error(),
		}).Error("failed to get session cookie")

		MakeResponse(w, http.StatusBadRequest,
			map[string]string{"message": "Invalid cookie"},
		)
		return
	}

	sh.Logger.WithFields(&logrus.Fields{
		"session_id": session.Value,
	}).Debug("attempting to logout session")

	if err := sh.LogoutUC.Logout(session.Value); err != nil {
		if err == model.ErrSessionNotFound {
			sh.Logger.WithFields(&logrus.Fields{
				"session_id": session.Value,
				"error":      err.Error(),
			}).Warn("session not found during logout")

			MakeResponse(w, http.StatusInternalServerError,
				map[string]string{"message": "session not found"},
			)
			return
		}
		if err == model.ErrGetSession {
			sh.Logger.WithFields(&logrus.Fields{
				"session_id": session.Value,
				"error":      err.Error(),
			}).Error("failed to get session during logout")

			MakeResponse(w, http.StatusInternalServerError,
				map[string]string{"message": "error getting session"},
			)
			return
		}
		if err == model.ErrDeleteSession {
			sh.Logger.WithFields(&logrus.Fields{
				"session_id": session.Value,
				"error":      err.Error(),
			}).Error("failed to delete session")

			MakeResponse(w, http.StatusInternalServerError,
				map[string]string{"message": "error deleting session"},
			)
			return
		}

		sh.Logger.WithFields(&logrus.Fields{
			"session_id": session.Value,
			"error":      err.Error(),
		}).Error("unknown logout error")

		MakeResponse(w, http.StatusInternalServerError,
			map[string]string{"message": "unknown logout error"},
		)
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

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		HttpOnly: true,
		Secure:   false,
		Expires:  time.Now().AddDate(-1, 0, 0),
		Path:     "/",
	})

	sh.Logger.WithFields(&logrus.Fields{
		"session_id": session.Value,
	}).Info("user logged out successfully")

	MakeResponse(w, http.StatusOK, map[string]string{"message": "Logged out"})
}

func (uh *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	uh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("DeleteUser request started")

	sanitizer := bluemonday.UGCPolicy()
	vars := mux.Vars(r)
	id := vars["id"]

	uh.Logger.WithFields(&logrus.Fields{
		"raw_user_id": id,
	}).Debug("received user ID for deletion")

	userId, err := strconv.Atoi(sanitizer.Sanitize(id))
	if err != nil {
		uh.Logger.WithFields(&logrus.Fields{
			"raw_user_id": id,
			"error":       err.Error(),
		}).Warn("invalid user ID format")

		MakeResponse(w, http.StatusBadRequest,
			map[string]string{"message": "Invalid user id"},
		)
		return
	}

	uh.Logger.WithFields(&logrus.Fields{
		"user_id": userId,
	}).Info("attempting to delete user")

	if err := uh.DeleteUserUC.DeleteUser(userId); err != nil {
		uh.Logger.WithFields(&logrus.Fields{
			"user_id": userId,
			"error":   err.Error(),
		}).Error("failed to delete user")

		MakeResponse(w, http.StatusInternalServerError,
			map[string]string{"message": "Error deleting user"},
		)
		return
	}

	uh.Logger.WithFields(&logrus.Fields{
		"user_id": userId,
	}).Info("user deleted successfully")

	MakeResponse(w, http.StatusOK,
		map[string]string{"message": fmt.Sprintf("User with ID %d deleted", userId)},
	)
}

func (gh *GetHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	gh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("GetProfile request started")

	userIDRaw := r.Context().Value(userIDKey)
	profileId, ok := userIDRaw.(uint32)
	if !ok {
		gh.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized profile access attempt")

		MakeResponse(w, http.StatusUnauthorized,
			map[string]string{"message": "You don't have access"},
		)
		return
	}

	gh.Logger.WithFields(&logrus.Fields{
		"profile_id": profileId,
	}).Debug("attempting to get profile")

	profile, err := gh.GetProfileUC.GetProfile(int(profileId))
	if err != nil {
		gh.Logger.WithFields(&logrus.Fields{
			"profile_id": profileId,
			"error":      err.Error(),
		}).Error("failed to get profile")

		MakeResponse(w, http.StatusInternalServerError,
			map[string]string{"message": fmt.Sprintf("Error getting profile: %v", err)},
		)
		return
	}

	gh.Logger.WithFields(&logrus.Fields{
		"profile_id": profileId,
	}).Info("profile retrieved successfully")

	MakeResponse(w, http.StatusOK, profile)
}

func (gh *GetHandler) GetProfiles(w http.ResponseWriter, r *http.Request) {
	gh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("GetProfiles request started")

	userIDRaw := r.Context().Value(userIDKey)
	profileId, ok := userIDRaw.(uint32)
	if !ok {
		gh.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized profiles access attempt")

		MakeResponse(w, http.StatusUnauthorized,
			map[string]string{"message": "You don't have access"},
		)
		return
	}

	gh.Logger.WithFields(&logrus.Fields{
		"requester_id": profileId,
	}).Debug("attempting to get profiles list")

	profiles, err := gh.GetProfilesUC.GetProfiles(int(profileId))
	if err != nil {
		gh.Logger.WithFields(&logrus.Fields{
			"requester_id": profileId,
			"error":        err.Error(),
		}).Error("failed to get profiles list")

		MakeResponse(w, http.StatusBadRequest,
			map[string]string{"message": fmt.Sprintf("Error getting profiles: %v", err)},
		)
		return
	}

	gh.Logger.WithFields(&logrus.Fields{
		"requester_id":   profileId,
		"profiles_count": len(profiles),
	}).Info("profiles list retrieved successfully")

	MakeResponse(w, http.StatusOK, profiles)
}

func (sh *StaticHandler) DeletePhoto(w http.ResponseWriter, r *http.Request) {
	sh.Logger.WithFields(&logrus.Fields{
		"method":       r.Method,
		"path":         r.URL.Path,
		"request_id":   r.Header.Get("request_id"),
		"ip":           r.RemoteAddr,
		"query_params": r.URL.Query(),
	}).Info("DeletePhoto request started")

	sanitizer := bluemonday.UGCPolicy()
	fileURL := sanitizer.Sanitize(r.URL.Query().Get("file_url"))

	sh.Logger.WithFields(&logrus.Fields{
		"raw_file_url":  r.URL.Query().Get("file_url"),
		"sanitized_url": fileURL,
	}).Debug("processing file URL")

	userIDRaw := r.Context().Value(userIDKey)
	user_id, ok := userIDRaw.(uint32)
	if !ok {
		sh.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized photo deletion attempt")

		MakeResponse(w, http.StatusUnauthorized,
			map[string]string{"message": "You don't have access"},
		)
		return
	}

	sh.Logger.WithFields(&logrus.Fields{
		"user_id":  user_id,
		"file_url": fileURL,
	}).Info("attempting to delete photo")

	err := sh.DeleteUC.DeleteImage(int(user_id), fileURL)
	if err != nil {
		sh.Logger.WithFields(&logrus.Fields{
			"user_id":  user_id,
			"file_url": fileURL,
			"error":    err.Error(),
		}).Error("failed to delete photo")

		MakeResponse(w, http.StatusInternalServerError,
			map[string]string{"message": fmt.Sprintf("Error deleting photo: %v", err)},
		)
		return
	}

	sh.Logger.WithFields(&logrus.Fields{
		"user_id":  user_id,
		"file_url": fileURL,
	}).Info("photo deleted successfully")

	MakeResponse(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Deleted photo %s for user %d", fileURL, user_id),
	})
}

func (qh *QueryHandler) GetActiveQueries(w http.ResponseWriter, r *http.Request) {
	qh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("GetActiveQueries request started")

	userIDRaw := r.Context().Value(userIDKey)
	user_id, ok := userIDRaw.(uint32)
	if !ok {
		qh.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized query access attempt")

		MakeResponse(w, http.StatusUnauthorized,
			map[string]string{"message": "You don't have access"},
		)
		return
	}

	qh.Logger.WithFields(&logrus.Fields{
		"user_id": user_id,
	}).Info("attempting to get active queries")

	queries, err := qh.GetActiveQueriesUC.GetActiveQueries(int32(user_id))
	if err != nil {
		qh.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Error("failed to get active queries")

		MakeResponse(w, http.StatusInternalServerError,
			map[string]string{"message": fmt.Sprintf("Error getting active queries: %v", err)},
		)
		return
	}

	qh.Logger.WithFields(&logrus.Fields{
		"user_id": user_id,
	}).Info("active queries retrieved successfully")

	MakeResponse(w, http.StatusOK, queries)
}

func (qh *QueryHandler) StoreUserAnswer(w http.ResponseWriter, r *http.Request) {
	qh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("SendUserAnswer request started")

	userIDRaw := r.Context().Value(userIDKey)
	user_id, ok := userIDRaw.(uint32)
	if !ok {
		qh.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized query access attempt")

		MakeResponse(w, http.StatusUnauthorized,
			map[string]string{"message": "You don't have access"},
		)
		return
	}

	var answer struct {
		Name   string `json:"name"`
		Score  int32  `json:"score"`
		Answer string `json:"answer"`
	}

	err := json.NewDecoder(r.Body).Decode(&answer)
	if err != nil {
		qh.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Error("failed to decode answer")

		MakeResponse(w, http.StatusBadRequest,
			map[string]string{"message": fmt.Sprintf("Error decoding answer: %v", err)},
		)
		return
	}

	qh.Logger.WithFields(&logrus.Fields{
		"user_id": user_id,
		"answer":  answer,
	}).Info("attempting to store user answer")

	err = qh.StoreUserAnswerUC.StoreUserAnswer(int32(user_id), answer.Name, answer.Score, answer.Answer)
	if err != nil {
		qh.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
			"answer":  answer,
			"error":   err.Error(),
		}).Error("failed to store user answer")

		MakeResponse(w, http.StatusInternalServerError,
			map[string]string{"message": fmt.Sprintf("Error storing user answer: %v", err)},
		)
		return
	}

	qh.Logger.WithFields(&logrus.Fields{
		"user_id": user_id,
		"answer":  answer,
	}).Info("user answer stored successfully")

	MakeResponse(w, http.StatusOK,
		map[string]string{"message": "User answer stored successfully"},
	)
}

func (qh *QueryHandler) GetAnswersForUser(w http.ResponseWriter, r *http.Request) {
	qh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("GetAnswersForUser request started")

	userIDRaw := r.Context().Value(userIDKey)
	user_id, ok := userIDRaw.(uint32)
	if !ok {
		qh.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized query access attempt")

		MakeResponse(w, http.StatusUnauthorized,
			map[string]string{"message": "You don't have access"},
		)
		return
	}

	qh.Logger.WithFields(&logrus.Fields{
		"user_id": user_id,
	}).Info("attempting to get answers for user")

	answers, err := qh.GetAnswersForUserUC.GetAnswersForUser(int32(user_id))
	if err != nil {
		qh.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Error("failed to get answers for user")

		MakeResponse(w, http.StatusInternalServerError,
			map[string]string{"message": fmt.Sprintf("Error getting answers for user: %v", err)},
		)
		return
	}

	qh.Logger.WithFields(&logrus.Fields{
		"user_id": user_id,
	}).Info("answers for user retrieved successfully")

	MakeResponse(w, http.StatusOK, answers)
}

func (qh *QueryHandler) GetAnswersForQuery(w http.ResponseWriter, r *http.Request) {
	qh.Logger.WithFields(&logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": r.Header.Get("request_id"),
		"ip":         r.RemoteAddr,
	}).Info("GetAnswersForQuery request started")

	userIDRaw := r.Context().Value(userIDKey)
	user_id, ok := userIDRaw.(uint32)
	if !ok {
		qh.Logger.WithFields(&logrus.Fields{
			"error": "missing or invalid userID in context",
		}).Warn("unauthorized query access attempt")

		MakeResponse(w, http.StatusUnauthorized,
			map[string]string{"message": "You don't have access"},
		)
		return
	}

	qh.Logger.WithFields(&logrus.Fields{
		"user_id": user_id,
	}).Info("attempting to get answers for query")

	answers, err := qh.GetAnswersForQueryUC.GetAnswersForQuery()
	if err != nil {
		qh.Logger.WithFields(&logrus.Fields{
			"user_id": user_id,
			"error":   err.Error(),
		}).Error("failed to get answers for query")

		MakeResponse(w, http.StatusInternalServerError,
			map[string]string{"message": fmt.Sprintf("Error getting answers for query: %v", err)},
		)
		return
	}

	qh.Logger.WithFields(&logrus.Fields{
		"user_id": user_id,
	}).Info("answers for query retrieved successfully")

	MakeResponse(w, http.StatusOK, answers)
}
