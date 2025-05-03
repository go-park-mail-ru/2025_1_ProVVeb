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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	sessionpb "github.com/go-park-mail-ru/2025_1_ProVVeb/auth_micro/proto"
	profilespb "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	querypb "github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/proto"
	userspb "github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/delivery"
)

func Run() {
	query_con, err := grpc.NewClient("query_micro:8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println(fmt.Errorf("not able to connect to query_micro: %v", err))
	}
	defer query_con.Close()

	auth_con, err := grpc.NewClient("auth_micro:8082", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println(fmt.Errorf("not able to connect to auth_micro: %v", err))
	}
	defer auth_con.Close()

	profiles_con, err := grpc.NewClient("profiles_micro:8083", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println(fmt.Errorf("not able to connect to profiles_micro: %v", err))
	}
	defer profiles_con.Close()

	users_con, err := grpc.NewClient("users_micro:8085", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println(fmt.Errorf("not able to connect to users_micro: %v", err))
	}
	defer users_con.Close()

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

	sessionHandler, err := NewSessionHandler(postgresClient, hasher, tokenValidator, validator, logger, auth_con)
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with sessionHandler: %v", err))
		return
	}

	usersHandler, err := NewUsersHandler(users_con, profiles_con, logger)
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with userHandler: %v", err))
		return
	}

	queryHandler, err := NewQueryHandler(query_con, logger)
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with queryHandler: %v", err))
		return
	}

	profilesHandler, err := NewProfilesHandler(profiles_con, logger)
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with profilesHandler: %v", err))
		return
	}

	r := mux.NewRouter()

	r.Use(RequestIDMiddleware)
	r.Use(PanicMiddleware(logger))
	r.Use(AccessLogMiddleware(logger))

	r.HandleFunc("/users", usersHandler.CreateUser).Methods("POST")
	r.HandleFunc("/users/login", sessionHandler.LoginUser).Methods("POST")
	r.HandleFunc("/users/logout", sessionHandler.LogoutUser).Methods("POST")

	usersSubrouter := r.PathPrefix("/users").Subrouter()
	usersSubrouter.Use(AuthWithCSRFMiddleware(tokenValidator, sessionHandler))
	usersSubrouter.Use(BodySizeLimitMiddleware(int64(model.Megabyte * model.MaxQuerySizeStr)))

	usersSubrouter.HandleFunc("/{id}", usersHandler.DeleteUser).Methods("DELETE")
	usersSubrouter.HandleFunc("/checkSession", sessionHandler.CheckSession).Methods("GET")
	usersSubrouter.HandleFunc("/getParams", usersHandler.GetUserParams).Methods("GET")

	profileSubrouter := r.PathPrefix("/profiles").Subrouter()
	profileSubrouter.Use(AuthWithCSRFMiddleware(tokenValidator, sessionHandler))
	profileSubrouter.Use(BodySizeLimitMiddleware(int64(model.Megabyte * model.MaxQuerySizeStr)))

	profileSubrouter.HandleFunc("/{id}", profilesHandler.GetProfile).Methods("GET")
	profileSubrouter.HandleFunc("", profilesHandler.GetProfiles).Methods("GET")
	profileSubrouter.HandleFunc("/like", profilesHandler.SetLike).Methods("POST")
	profileSubrouter.HandleFunc("/match/{id}", profilesHandler.GetMatches).Methods("GET")
	profileSubrouter.HandleFunc("/update", profilesHandler.UpdateProfile).Methods("POST")

	photoSubrouter := r.PathPrefix("/profiles").Subrouter()
	photoSubrouter.Use(AuthWithCSRFMiddleware(tokenValidator, sessionHandler))
	photoSubrouter.Use(BodySizeLimitMiddleware(int64(model.Megabyte * model.MaxQuerySizePhoto)))

	photoSubrouter.HandleFunc("/uploadPhoto", profilesHandler.UploadPhoto).Methods("POST")
	photoSubrouter.HandleFunc("/deletePhoto", profilesHandler.DeletePhoto).Methods("DELETE")

	querySubrouter := r.PathPrefix("/queries").Subrouter()
	querySubrouter.Use(AuthWithCSRFMiddleware(tokenValidator, sessionHandler))
	querySubrouter.Use(BodySizeLimitMiddleware(int64(model.Megabyte * model.MaxQuerySizeStr)))

	querySubrouter.HandleFunc("/getActive", queryHandler.GetActiveQueries).Methods("GET")
	querySubrouter.HandleFunc("/sendResp", queryHandler.StoreUserAnswer).Methods("POST")
	querySubrouter.HandleFunc("/getForUser", queryHandler.GetAnswersForUser).Methods("GET")
	querySubrouter.HandleFunc("/getForQuery", queryHandler.GetAnswersForQuery).Methods("GET")

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

func NewQueryHandler(
	conn *grpc.ClientConn,
	logger *logger.LogrusLogger,
) (*QueryHandler, error) {
	client := querypb.NewQueryServiceClient(conn)

	GetActive, err := usecase.NewGetActiveQueriesUseCase(client, logger)
	if err != nil {
		return nil, err
	}

	StoreAnswers, err := usecase.NewStoreUserAnswer(client, logger)
	if err != nil {
		return nil, err
	}

	GetAnswersForUser, err := usecase.NewGetAnswersForUserUseCase(client, logger)
	if err != nil {
		return nil, err
	}

	GetAnswersForQuery, err := usecase.NewGetAnswersForQueryUseCase(client, logger)
	if err != nil {
		return nil, err
	}
	return &QueryHandler{
		GetActiveQueriesUC:   *GetActive,
		StoreUserAnswerUC:    *StoreAnswers,
		GetAnswersForUserUC:  *GetAnswersForUser,
		GetAnswersForQueryUC: *GetAnswersForQuery,
		Logger:               logger,
	}, nil
}

func NewProfilesHandler(
	conn *grpc.ClientConn,
	logger *logger.LogrusLogger,
) (*ProfilesHandler, error) {
	client := profilespb.NewProfilesServiceClient(conn)

	DeleteImage, err := usecase.NewDeleteStaticUseCase(client, logger)
	if err != nil {
		return nil, err
	}

	GetProfileImages, err := usecase.NewGetUserPhotoUseCase(client, logger)
	if err != nil {
		return nil, err
	}

	GetProfileMatches, err := usecase.NewGetProfileMatchesUseCase(client, logger)
	if err != nil {
		return nil, err
	}

	GetProfile, err := usecase.NewGetProfileUseCase(client, logger)
	if err != nil {
		return nil, err
	}

	GetProfiles, err := usecase.NewGetProfilesForUserUseCase(client, logger)
	if err != nil {
		return nil, err
	}

	SetProfilesLike, err := usecase.NewProfileSetLikeUseCase(client, logger)
	if err != nil {
		return nil, err
	}

	UpdateProfile, err := usecase.NewProfileUpdateUseCase(client, logger)
	if err != nil {
		return nil, err
	}

	UpdateProfileImages, err := usecase.NewStaticUploadUseCase(client, logger)
	if err != nil {
		return nil, err
	}

	return &ProfilesHandler{
		DeleteImageUC:         *DeleteImage,
		GetProfileImagesUC:    *GetProfileImages,
		GetProfileMatchesUC:   *GetProfileMatches,
		GetProfileUC:          *GetProfile,
		GetProfilesUC:         *GetProfiles,
		SetProfilesLikeUC:     *SetProfilesLike,
		UpdateProfileUC:       *UpdateProfile,
		UpdateProfileImagesUC: *UpdateProfileImages,
		Logger:                logger,
	}, nil
}

func NewSessionHandler(
	userRepo repository.UserRepository,
	hasher repository.PasswordHasher,
	token repository.JwtTokenizer,
	validator repository.UserParamsValidator,
	logger *logger.LogrusLogger,
	conn *grpc.ClientConn,
) (*SessionHandler, error) {

	client := sessionpb.NewSessionServiceClient(conn)

	loginUC, err := usecase.NewUserLogInUseCase(
		userRepo,
		hasher,
		token,
		validator,
		client,
		logger,
	)
	if err != nil {
		return &SessionHandler{}, err
	}

	checkSessionUC, err := usecase.NewUserCheckSessionUseCase(
		client,
		logger,
	)
	if err != nil {
		return &SessionHandler{}, err
	}

	logoutUC, err := usecase.NewUserLogOutUseCase(
		userRepo,
		client,
		logger,
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

func NewUsersHandler(
	userConn *grpc.ClientConn,
	profilesConn *grpc.ClientConn,
	logger *logger.LogrusLogger,
) (*UserHandler, error) {
	usersClient := userspb.NewUsersServiceClient(userConn)
	profilesClient := profilespb.NewProfilesServiceClient(profilesConn)

	SignupUC, err := usecase.NewUserSignUpUseCase(usersClient, profilesClient, logger)
	if err != nil {
		return &UserHandler{}, err
	}
	DeleteUserUC, err := usecase.NewUserDeleteUseCase(usersClient, profilesClient, logger)
	if err != nil {
		return &UserHandler{}, err
	}

	GetUserParamsUC, err := usecase.NewUserGetParamsUseCase(usersClient, logger)
	if err != nil {
		return &UserHandler{}, err
	}

	return &UserHandler{
		SignupUC:     *SignupUC,
		DeleteUserUC: *DeleteUserUC,
		GetParamsUC:  *GetUserParamsUC,
		Logger:       logger,
	}, nil
}
