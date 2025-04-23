package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/usecase"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func Run() {
	logger, err := logger.NewLogrusLogger("/backend/logs/access.log")
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}

	tokenValidator, _ := repository.NewJwtToken(string(model.Key))
	postgresClient, err := repository.NewUserRepo()
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with postgresClient: %v", err))
		return
	}
	defer postgresClient.CloseRepo()

	redisClient, err := repository.NewSessionRepo()
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with redisClient: %v", err))
		return
	}
	defer redisClient.CloseRepo()

	staticClient, err := repository.NewStaticRepo()
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with minio: %v", err))
		return
	}

	hasher, err := repository.NewPassHasher()
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with hasher: %v", err))
		return
	}

	validator, err := repository.NewUParamsValidator()
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with validator: %v", err))
		return
	}

	getHandler, err := NewGetHandler(postgresClient, staticClient, logger)
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with getHandler: %v", err))
		return
	}

	sessionHandler, err := NewSessionHandler(postgresClient, redisClient, hasher, tokenValidator, validator, logger)
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with sessionHandler: %v", err))
		return
	}

	profileHandler, err := NewProfileHandler(postgresClient, staticClient, logger)
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with profileHandler: %v", err))
		return
	}

	userHandler, err := NewUserHandler(postgresClient, staticClient, hasher, validator, logger)
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with userHandler: %v", err))
		return
	}

	staticHandler, err := NewStaticHandler(postgresClient, staticClient, logger)
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with staticHandler: %v", err))
		return
	}

	r := mux.NewRouter()

	r.Use(RequestIDMiddleware)
	r.Use(PanicMiddleware(logger))
	r.Use(AccessLogMiddleware(logger))

	r.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	r.HandleFunc("/users/login", sessionHandler.LoginUser).Methods("POST")
	r.HandleFunc("/users/logout", sessionHandler.LogoutUser).Methods("POST")
	r.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")
	r.HandleFunc("/users/checkSession", sessionHandler.CheckSession).Methods("GET")

	usersSubrouter := r.PathPrefix("/users").Subrouter()

	usersSubrouter.HandleFunc("/{id}", userHandler.DeleteUser).Methods("DELETE")
	usersSubrouter.HandleFunc("/checkSession", sessionHandler.CheckSession).Methods("GET")
	usersSubrouter.Use(AuthWithCSRFMiddleware(tokenValidator, sessionHandler))

	profileSubrouter := r.PathPrefix("/profiles").Subrouter()

	profileSubrouter.Use(AuthWithCSRFMiddleware(tokenValidator, sessionHandler))
	profileSubrouter.Use(BodySizeLimitMiddleware(int64(model.Megabyte * model.MaxQuerySizeStr)))

	profileSubrouter.HandleFunc("/{id}", getHandler.GetProfile).Methods("GET")
	profileSubrouter.HandleFunc("", getHandler.GetProfiles).Methods("GET")
	profileSubrouter.HandleFunc("/like", profileHandler.SetLike).Methods("POST")
	profileSubrouter.HandleFunc("/match/{id}", profileHandler.GetMatches).Methods("GET")
	profileSubrouter.HandleFunc("/update", profileHandler.UpdateProfile).Methods("POST")

	photoSubrouter := r.PathPrefix("/profiles").Subrouter()

	photoSubrouter.Use(AuthWithCSRFMiddleware(tokenValidator, sessionHandler))
	photoSubrouter.Use(BodySizeLimitMiddleware(int64(model.Megabyte * model.MaxQuerySizePhoto)))

	photoSubrouter.HandleFunc("/uploadPhoto", staticHandler.UploadPhoto).Methods("POST")
	photoSubrouter.HandleFunc("/deletePhoto", staticHandler.DeletePhoto).Methods("DELETE")

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://213.219.214.83:8000", "http://localhost:8000"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT"},
		AllowedHeaders:   []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	handler := corsMiddleware.Handler(r)

	server := http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("starting server at :8080")
	fmt.Println(fmt.Errorf("server ended with error: %v", server.ListenAndServe()))
}

func NewGetHandler(
	userRepo repository.UserRepository,
	staticRepo repository.StaticRepository,
	logger *logger.LogrusLogger,
) (*GetHandler, error) {

	getProfileUC, err := usecase.NewGetProfileUseCase(userRepo, staticRepo)
	if err != nil {
		return nil, err
	}

	getProfilesForUserUC, err := usecase.NewGetProfilesForUserUseCase(userRepo, staticRepo)
	if err != nil {
		return nil, err
	}

	getUserPhotoUC, err := usecase.NewGetUserPhotoUseCase(userRepo, staticRepo)
	if err != nil {
		return nil, err
	}

	return &GetHandler{
		GetProfileUC:    *getProfileUC,
		GetProfilesUC:   *getProfilesForUserUC,
		GetProfileImage: *getUserPhotoUC,
		Logger:          logger,
	}, nil
}

func NewSessionHandler(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	hasher repository.PasswordHasher,
	token repository.JwtTokenizer,
	validator repository.UserParamsValidator,
	logger *logger.LogrusLogger,
) (*SessionHandler, error) {
	loginUC, err := usecase.NewUserLogInUseCase(
		userRepo,
		sessionRepo,
		hasher,
		token,
		validator,
	)
	if err != nil {
		return &SessionHandler{}, err
	}

	checkSessionUC, err := usecase.NewUserCheckSessionUseCase(
		sessionRepo,
	)
	if err != nil {
		return &SessionHandler{}, err
	}

	logoutUC, err := usecase.NewUserLogOutUseCase(
		userRepo,
		sessionRepo,
	)
	if err != nil {
		return &SessionHandler{}, err
	}

	return &SessionHandler{
		LoginUC:        *loginUC,
		CheckSessionUC: *checkSessionUC,
		LogoutUC:       *logoutUC,
		Logger:         logger,
	}, nil
}

func NewProfileHandler(
	userRepo repository.UserRepository,
	staticRepo repository.StaticRepository,
	logger *logger.LogrusLogger,
) (*ProfileHandler, error) {
	likeUC, err := usecase.NewProfileSetLikeUseCase(
		userRepo,
	)
	if err != nil {
		return &ProfileHandler{}, err
	}
	matchUC, err := usecase.NewGetProfileMatchesUseCase(
		userRepo,
	)
	if err != nil {
		return &ProfileHandler{}, err
	}
	updateUC, err := usecase.NewProfileUpdateUseCase(
		userRepo,
	)
	if err != nil {
		return &ProfileHandler{}, err
	}
	getProfileUC, err := usecase.NewGetProfileUseCase(
		userRepo,
		staticRepo,
	)
	if err != nil {
		return &ProfileHandler{}, err
	}
	getUserPhotoUC, err := usecase.NewGetUserPhotoUseCase(
		userRepo,
		staticRepo,
	)
	if err != nil {
		return &ProfileHandler{}, err
	}
	return &ProfileHandler{
		LikeUC:            *likeUC,
		MatchUC:           *matchUC,
		UpdateUC:          *updateUC,
		GetProfileUC:      *getProfileUC,
		GetProfileImageUC: *getUserPhotoUC,
		Logger:            logger,
	}, nil
}

func NewStaticHandler(
	userRepo repository.UserRepository,
	staticRepo repository.StaticRepository,
	logger *logger.LogrusLogger,
) (*StaticHandler, error) {
	uploadUC, err := usecase.NewStaticUploadUseCase(
		userRepo,
		staticRepo,
	)
	if err != nil {
		return &StaticHandler{}, err
	}

	deleteUC, err := usecase.NewDeleteStaticUseCase(
		userRepo,
		staticRepo,
	)
	if err != nil {
		return &StaticHandler{}, err
	}

	return &StaticHandler{
		UploadUC: *uploadUC,
		DeleteUC: *deleteUC,
		Logger:   logger,
	}, nil
}

func NewUserHandler(
	userRepo repository.UserRepository,
	staticRepo repository.StaticRepository,
	hasher repository.PasswordHasher,
	validator repository.UserParamsValidator,
	logger *logger.LogrusLogger,
) (*UserHandler, error) {
	signupUC, err := usecase.NewUserSignUpUseCase(
		userRepo,
		staticRepo,
		hasher,
		validator,
	)
	if err != nil {
		return &UserHandler{}, err
	}
	deleteUserUC, err := usecase.NewUserDeleteUseCase(
		userRepo,
	)
	if err != nil {
		return &UserHandler{}, err
	}

	return &UserHandler{
		SignupUC:     *signupUC,
		DeleteUserUC: *deleteUserUC,
		Logger:       logger,
	}, nil
}
