package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/logger"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/usecase"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	sessionpb "github.com/go-park-mail-ru/2025_1_ProVVeb/auth_micro/proto"
	profilespb "github.com/go-park-mail-ru/2025_1_ProVVeb/profiles_micro/delivery"
	querypb "github.com/go-park-mail-ru/2025_1_ProVVeb/query_micro/proto"
	userspb "github.com/go-park-mail-ru/2025_1_ProVVeb/users_micro/delivery"
)

func Run() {
	registry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = registry
	prometheus.DefaultGatherer = registry

	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	grpcMetrics := grpc_prometheus.NewClientMetrics()
	registry.MustRegister(grpcMetrics)

	grpcOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcMetrics.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(grpcMetrics.StreamClientInterceptor()),
	}

	queryCon, err := grpc.NewClient("query_micro:8081", grpcOpts...)
	if err != nil {
		fmt.Println(fmt.Errorf("not able to connect to query_micro: %v", err))
	}
	defer queryCon.Close()

	authCon, err := grpc.NewClient("auth_micro:8082", grpcOpts...)
	if err != nil {
		fmt.Println(fmt.Errorf("not able to connect to auth_micro: %v", err))
	}
	defer authCon.Close()

	profilesCon, err := grpc.NewClient("profiles_micro:8083", grpcOpts...)
	if err != nil {
		fmt.Println(fmt.Errorf("not able to connect to profiles_micro: %v", err))
	}
	defer profilesCon.Close()

	usersCon, err := grpc.NewClient("users_micro:8085", grpcOpts...)
	if err != nil {
		fmt.Println(fmt.Errorf("not able to connect to users_micro: %v", err))
	}
	defer usersCon.Close()

	logger, err := logger.NewLogrusLogger("/backend/logs/access.log")
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}

	chatClient, err := repository.NewChatRepo()
	if err != nil {
		fmt.Printf("Failed to initialize chat repo: %v\n", err)
		return
	}

	notifClient, err := repository.NewNotificationsRepo()
	if err != nil {
		fmt.Printf("Failed to initialize notificatins repo: %v\n", err)
		return
	}

	complaintClient, err := repository.NewComplaintRepo()
	if err != nil {
		fmt.Printf("Failed to initialize complaint repo: %v\n", err)
		return
	}

	subClient, err := repository.NewSubRepo()
	if err != nil {
		fmt.Printf("Failed to initialize complaint repo: %v\n", err)
		return
	}

	tokenValidator, _ := repository.NewJwtToken(string(model.Key))

	hasher, err := repository.NewPassHasher()
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with hasher: %v", err))
		return
	}

	sessionHandler, err := NewSessionHandler(hasher, tokenValidator, logger, authCon, usersCon)
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with sessionHandler: %v", err))
		return
	}

	usersHandler, err := NewUsersHandler(usersCon, profilesCon, logger)
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with userHandler: %v", err))
		return
	}

	queryHandler, err := NewQueryHandler(queryCon, usersCon, logger)
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with queryHandler: %v", err))
		return
	}

	messageHandler, err := NewMessageHandler(chatClient, notifClient, chatClient.Client, logger)
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with queryHandler: %v", err))
		return
	}

	profilesHandler, err := NewProfilesHandler(profilesCon, notifClient, usersCon, notifClient.Client.(*redis.Client), logger)
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with profilesHandler: %v", err))
		return
	}

	complaintHandler, err := NewComplaintHandler(complaintClient, usersCon, logger)
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with complaintHandler: %v", err))
		return
	}

	notificationHandler, err := NewNotificationHandler(notifClient, notifClient.Client.(*redis.Client), logger)
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with notificationHandler: %v", err))
		return
	}

	subscripHandler, err := NewSubHandler(subClient, logger)
	if err != nil {
		fmt.Println(fmt.Errorf("not able to work with notificationHandler: %v", err))
		return
	}

	r := mux.NewRouter()

	metricsMiddleware := NewMetricsMiddleware(MetricsMiddlewareConfig{Registry: registry})
	r.Use(metricsMiddleware)
	r.Use(RequestIDMiddleware)
	r.Use(PanicMiddleware(logger))
	r.Use(AccessLogMiddleware(logger))

	r.HandleFunc("/users", usersHandler.CreateUser).Methods("POST")
	r.HandleFunc("/users/login", sessionHandler.LoginUser).Methods("POST")
	r.HandleFunc("/users/logout", sessionHandler.LogoutUser).Methods("POST")

	usersSubrouter := r.PathPrefix("/users").Subrouter()
	usersSubrouter.Use(AuthWithCSRFMiddleware(tokenValidator, sessionHandler, usersHandler))
	usersSubrouter.Use(BodySizeLimitMiddleware(int64(model.Megabyte * model.MaxQuerySizeStr)))

	usersSubrouter.HandleFunc("/{id}", usersHandler.DeleteUser).Methods("DELETE")
	usersSubrouter.HandleFunc("/checkSession", sessionHandler.CheckSession).Methods("GET")
	usersSubrouter.HandleFunc("/getParams", usersHandler.GetUserParams).Methods("GET")

	profileSubrouter := r.PathPrefix("/profiles").Subrouter()
	profileSubrouter.Use(AuthWithCSRFMiddleware(tokenValidator, sessionHandler, usersHandler))
	profileSubrouter.Use(BodySizeLimitMiddleware(int64(model.Megabyte * model.MaxQuerySizeStr)))

	profileSubrouter.HandleFunc("", profilesHandler.GetProfiles).Methods("GET")
	profileSubrouter.HandleFunc("/like", profilesHandler.SetLike).Methods("POST")
	profileSubrouter.HandleFunc("/match/{id}", profilesHandler.GetMatches).Methods("GET")
	profileSubrouter.HandleFunc("/update", profilesHandler.UpdateProfile).Methods("POST")
	profileSubrouter.HandleFunc("/search", profilesHandler.SearchProfiles).Methods("POST")
	profileSubrouter.HandleFunc("/recommendations", profilesHandler.GetRecommendations).Methods("GET")
	profileSubrouter.HandleFunc("/getStatistics", profilesHandler.GetStatistics).Methods("GET")
	profileSubrouter.HandleFunc("/{id}", profilesHandler.GetProfile).Methods("GET")

	photoSubrouter := r.PathPrefix("/profiles").Subrouter()
	photoSubrouter.Use(AuthWithCSRFMiddleware(tokenValidator, sessionHandler, usersHandler))
	photoSubrouter.Use(BodySizeLimitMiddleware(int64(model.Megabyte * model.MaxQuerySizePhoto)))

	photoSubrouter.HandleFunc("/uploadPhoto", profilesHandler.UploadPhoto).Methods("POST")
	photoSubrouter.HandleFunc("/deletePhoto", profilesHandler.DeletePhoto).Methods("DELETE")

	querySubrouter := r.PathPrefix("/queries").Subrouter()
	querySubrouter.Use(AuthWithCSRFMiddleware(tokenValidator, sessionHandler, usersHandler))
	querySubrouter.Use(BodySizeLimitMiddleware(int64(model.Megabyte * model.MaxQuerySizeStr)))

	querySubrouter.HandleFunc("/getActive", queryHandler.GetActiveQueries).Methods("GET")
	querySubrouter.HandleFunc("/sendResp", queryHandler.StoreUserAnswer).Methods("POST")
	querySubrouter.HandleFunc("/getForUser", queryHandler.GetAnswersForUser).Methods("GET")
	querySubrouter.HandleFunc("/getForQuery", queryHandler.GetAnswersForQuery).Methods("GET")
	querySubrouter.HandleFunc("/findQuery", queryHandler.FindQuery).Methods("POST")

	querySubrouter.HandleFunc("/deleteAnswer", queryHandler.DeleteQuery).Methods("POST")
	querySubrouter.HandleFunc("/getStatistics", queryHandler.GetStatistics).Methods("POST")

	messageSubrouter := r.PathPrefix("/chats").Subrouter()
	messageSubrouter.Use(AuthWithCSRFMiddleware(tokenValidator, sessionHandler, usersHandler))
	messageSubrouter.Use(BodySizeLimitMiddleware(int64(model.Megabyte * model.MaxQuerySizeStr)))

	messageSubrouter.HandleFunc("", messageHandler.GetChats).Methods("GET")
	messageSubrouter.HandleFunc("/create", messageHandler.CreateChat).Methods("POST")
	messageSubrouter.HandleFunc("/delete", messageHandler.DeleteChat).Methods("DELETE")

	wsRouter := r.PathPrefix("/chats").Subrouter()
	wsRouter.Use(AuthWithCSRFMiddleware(tokenValidator, sessionHandler, usersHandler))
	wsRouter.Use(BodySizeLimitMiddleware(int64(model.Megabyte * model.MaxQuerySizeStr)))

	wsRouter.HandleFunc("/{chat_id}", messageHandler.HandleChat).Methods("GET")

	notificationsSubrouter := r.PathPrefix("/notifications").Subrouter()
	notificationsSubrouter.Use(AuthWithCSRFMiddleware(tokenValidator, sessionHandler, usersHandler))
	notificationsSubrouter.Use(BodySizeLimitMiddleware(int64(model.Megabyte * model.MaxQuerySizeStr)))

	notificationsSubrouter.HandleFunc("", notificationHandler.GetNotifications).Methods("GET")

	ComplaintSubrouter := r.PathPrefix("/complaints").Subrouter()
	ComplaintSubrouter.Use(AuthWithCSRFMiddleware(tokenValidator, sessionHandler, usersHandler))
	ComplaintSubrouter.Use(BodySizeLimitMiddleware(int64(model.Megabyte * model.MaxQuerySizeStr)))

	ComplaintSubrouter.HandleFunc("/create", complaintHandler.CreateComplaint).Methods("POST")
	ComplaintSubrouter.HandleFunc("/get", complaintHandler.GetComplaints).Methods("GET")
	ComplaintSubrouter.HandleFunc("/find", complaintHandler.FindComplaint).Methods("POST")
	ComplaintSubrouter.HandleFunc("/delete", complaintHandler.DeleteComplaint).Methods("DELETE")
	ComplaintSubrouter.HandleFunc("/handle", complaintHandler.HandleComplaint).Methods("POST")
	ComplaintSubrouter.HandleFunc("/getStatistics", complaintHandler.GetStatistics).Methods("POST")

	subscriptionSubrouter := r.PathPrefix("/subscription").Subrouter()
	subscriptionSubrouter.Use(AuthWithCSRFMiddleware(tokenValidator, sessionHandler, usersHandler))
	subscriptionSubrouter.Use(BodySizeLimitMiddleware(int64(model.Megabyte * model.MaxQuerySizeStr)))

	subscriptionSubrouter.HandleFunc("", subscripHandler.AddSubscription).Methods("POST")
	subscriptionSubrouter.HandleFunc("/changeborder", subscripHandler.ChangeBorder).Methods("POST")

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://213.219.214.83:8000",
			"http://localhost:8000",
			"http://localhost",
			"http://beameye.ru:8000",
			"http://beameye.ru",
		},
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

	rmetrics := mux.NewRouter()
	rmetrics.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))

	go http.ListenAndServe(":8099", rmetrics)
	go func() {
		for {
			updateSyscallMetrics()
			time.Sleep(5 * time.Second)
		}
	}()

	fmt.Println("starting server at :8080")
	fmt.Println(fmt.Errorf("server ended with error: %v", server.ListenAndServe()))
}

func NewComplaintHandler(
	complRepo repository.ComplaintRepository,
	admin_conn *grpc.ClientConn,
	logger *logger.LogrusLogger,
) (*ComplaintHandler, error) {
	admin_client := userspb.NewUsersServiceClient(admin_conn)

	GetComplaints, err := usecase.NewGetComplaintUseCase(complRepo, logger)
	if err != nil {
		return nil, err
	}

	CreateComplate, err := usecase.NewCreateComplaintUseCase(complRepo, logger)
	if err != nil {
		return nil, err
	}

	FindComplaintUC, err := usecase.NewFindComplaintUseCase(complRepo, logger)
	if err != nil {
		return nil, err
	}

	DeleteComplaintsUC, err := usecase.NewDeleteComplaintUseCase(complRepo, logger)
	if err != nil {
		return nil, err
	}

	HandleComplaintUC, err := usecase.NewHandleComplaintUseCase(complRepo, logger)
	if err != nil {
		return nil, err
	}

	GetStatisticsUC, err := usecase.NewGetStatisticsComplUseCase(complRepo, logger)
	if err != nil {
		return nil, err
	}

	GetAdminUC, err := usecase.NewGetAdminUseCase(admin_client, logger)
	if err != nil {
		return nil, err
	}

	return &ComplaintHandler{
		GetComplaintsUC:    *GetComplaints,
		CreateComplateUC:   *CreateComplate,
		FindCompaintUC:     *FindComplaintUC,
		DeleteComplaintsUC: *DeleteComplaintsUC,
		HandleComplaintUC:  *HandleComplaintUC,
		GetStatisticsUC:    *GetStatisticsUC,
		GetAdminUC:         *GetAdminUC,
		Logger:             logger,
	}, nil
}

func NewQueryHandler(
	query_conn *grpc.ClientConn,
	admin_conn *grpc.ClientConn,
	logger *logger.LogrusLogger,
) (*QueryHandler, error) {
	query_client := querypb.NewQueryServiceClient(query_conn)
	admin_client := userspb.NewUsersServiceClient(admin_conn)

	GetActive, err := usecase.NewGetActiveQueriesUseCase(query_client, logger)
	if err != nil {
		return nil, err
	}

	StoreAnswers, err := usecase.NewStoreUserAnswer(query_client, logger)
	if err != nil {
		return nil, err
	}

	GetAnswersForUser, err := usecase.NewGetAnswersForUserUseCase(query_client, logger)
	if err != nil {
		return nil, err
	}

	GetAnswersForQuery, err := usecase.NewGetAnswersForQueryUseCase(query_client, logger)
	if err != nil {
		return nil, err
	}

	GetAdminUC, err := usecase.NewGetAdminUseCase(admin_client, logger)
	if err != nil {
		return nil, err
	}

	FindQueryUC, err := usecase.NewFindQueryUseCase(query_client, logger)
	if err != nil {
		return nil, err
	}

	DeleteQueryUC, err := usecase.NewDeleteAnswerUseCase(query_client, logger)
	if err != nil {
		return nil, err
	}

	GetStatisticsUC, err := usecase.NewGetStatisticsUseCase(query_client, logger)
	if err != nil {
		return nil, err
	}

	return &QueryHandler{
		GetActiveQueriesUC:   *GetActive,
		StoreUserAnswerUC:    *StoreAnswers,
		GetAnswersForUserUC:  *GetAnswersForUser,
		GetAnswersForQueryUC: *GetAnswersForQuery,
		GetAdminUC:           *GetAdminUC,
		FindQueryUC:          *FindQueryUC,
		DeleteQueryUC:        *DeleteQueryUC,
		GetStatisticsUC:      *GetStatisticsUC,
		Logger:               logger,
	}, nil
}

func NewMessageHandler(
	messageRepo repository.ChatRepository,
	notifrepo repository.NotificationsRepository,
	Subscriber *redis.Client,
	logger *logger.LogrusLogger,
) (*MessageHandler, error) {

	getChatsUC, err := usecase.NewGetChatsUseCase(messageRepo, logger)
	if err != nil {
		return nil, err
	}
	createChatsUC, err := usecase.NewCreateChatUseCase(messageRepo, logger)
	if err != nil {
		return nil, err
	}

	deleteChatsUC, err := usecase.NewDeleteChatUseCase(messageRepo, logger)
	if err != nil {
		return nil, err
	}

	getMessages, err := usecase.NewGetMessagesUseCase(messageRepo, logger)
	if err != nil {
		return nil, err
	}

	deleteMessage, err := usecase.NewDeleteMessageUseCase(messageRepo, logger)
	if err != nil {
		return nil, err
	}
	createMessageUC, err := usecase.NewCreateMessagesUseCase(messageRepo, logger)
	if err != nil {
		return nil, err
	}
	getMessagesFromCacheUC, err := usecase.NewGetMessagesFromCacheUseCase(messageRepo, logger)
	if err != nil {
		return nil, err
	}
	updateMessageStatusUC, err := usecase.NewUpdateMessageStatusUseCase(messageRepo, logger)
	if err != nil {
		return nil, err
	}

	getParticipantsUC, err := usecase.NewGetChatParticipantsUseCase(messageRepo, logger)
	if err != nil {
		return nil, err
	}

	AddNotification, err := usecase.NewAddNotificationUseCase(notifrepo, logger)
	if err != nil {
		return nil, err
	}

	return &MessageHandler{
		GetParticipantsUC:      *getParticipantsUC,
		GetChatsUC:             *getChatsUC,
		CreateChatUC:           *createChatsUC,
		DeleteChatUC:           *deleteChatsUC,
		GetMessagesUC:          *getMessages,
		DeleteMessageUC:        *deleteMessage,
		CreateMessagesUC:       *createMessageUC,
		GetMessagesFromCacheUC: *getMessagesFromCacheUC,
		UpdateMessageStatusUC:  *updateMessageStatusUC,
		AddNotificationUC:      *AddNotification,
		Subscriber:             Subscriber,
		Logger:                 logger,
	}, nil
}

func NewProfilesHandler(
	conn *grpc.ClientConn,
	notifrepo repository.NotificationsRepository,
	admin_conn *grpc.ClientConn,
	Subscriber *redis.Client,
	logger *logger.LogrusLogger,
) (*ProfilesHandler, error) {
	client := profilespb.NewProfilesServiceClient(conn)
	admin_client := userspb.NewUsersServiceClient(admin_conn)

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

	SearchProfile, err := usecase.NewSearchProfilesUseCase(client, logger)
	if err != nil {
		return nil, err
	}

	AddNotification, err := usecase.NewAddNotificationUseCase(notifrepo, logger)
	if err != nil {
		return nil, err
	}

	GetAdmin, err := usecase.NewGetAdminUseCase(admin_client, logger)
	if err != nil {
		return nil, err
	}

	GetRecommendations, err := usecase.NewGetRecommendationsUseCase(client, logger)
	if err != nil {
		return nil, err
	}

	GetProfileStats, err := usecase.NewGetProfileStatsUseCase(client, logger)
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
		Subscriber:            Subscriber,
		AddNotificationUC:     *AddNotification,
		GetRecommendationsUC:  *GetRecommendations,
		GetProfileStatsUC:     *GetProfileStats,
		SearchProfileUC:       *SearchProfile,
		GetAdminUC:            *GetAdmin,
		Logger:                logger,
	}, nil
}

func NewSessionHandler(
	hasher repository.PasswordHasher,
	token repository.JwtTokenizer,
	logger *logger.LogrusLogger,
	conn *grpc.ClientConn,
	userConn *grpc.ClientConn,
) (*SessionHandler, error) {

	client := sessionpb.NewSessionServiceClient(conn)

	userClient := userspb.NewUsersServiceClient(userConn)

	loginUC, err := usecase.NewUserLogInUseCase(
		hasher,
		token,
		userClient,
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

	GetPremiumUC, err := usecase.NewGetPremiumUseCase(usersClient, logger)
	if err != nil {
		return &UserHandler{}, err
	}

	return &UserHandler{
		SignupUC:     *SignupUC,
		DeleteUserUC: *DeleteUserUC,
		GetParamsUC:  *GetUserParamsUC,
		GetPremiumUC: *GetPremiumUC,
		Logger:       logger,
	}, nil
}

func NewNotificationHandler(
	notifRepo repository.NotificationsRepository,
	Subscriber *redis.Client,
	logger *logger.LogrusLogger,
) (*NotificationsHandler, error) {
	GetNotifications, err := usecase.NewGetNotificationsUseCase(notifRepo, logger)
	if err != nil {
		return nil, err
	}

	UpdateNotifications, err := usecase.NewUpdateNotificationStatusUseCase(notifRepo, logger)
	if err != nil {
		return nil, err
	}

	DeleteNotifications, err := usecase.NewDeleteNotificationUseCase(notifRepo, logger)
	if err != nil {
		return nil, err
	}

	GetCurrentNotification, err := usecase.NewGetCurrentNotificationsUseCase(notifRepo, logger)
	if err != nil {
		return nil, err
	}

	return &NotificationsHandler{
		GetNotificationsUC:         *GetNotifications,
		UpdateNotificationStatusUC: *UpdateNotifications,
		DeleteNotificationUC:       *DeleteNotifications,
		GetCurrentNotificationsUC:  *GetCurrentNotification,
		Subscriber:                 Subscriber,
		Logger:                     logger,
	}, nil

}

func NewSubHandler(
	subClient repository.SubsriptionRepository,
	logger *logger.LogrusLogger,
) (*SubHandler, error) {
	AddSubscription, err := usecase.NewAddSubscriptionUseCase(subClient, logger)
	if err != nil {
		return nil, err
	}
	UpdateBorder, err := usecase.NewUpdateBorderUseCase(subClient, logger)
	if err != nil {
		return nil, err
	}
	return &SubHandler{
		AddSubUC:       *AddSubscription,
		UpdateBorderUC: *UpdateBorder,
		Logger:         logger,
	}, nil
}
