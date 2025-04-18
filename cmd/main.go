package main

import (
	"fmt"
	"net/http"
	"time"

	handlers "github.com/go-park-mail-ru/2025_1_ProVVeb/delivery"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/usecase"

	"github.com/gorilla/mux"

	"github.com/rs/cors"
)

func main() {
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

	r := mux.NewRouter()

	getHandler := &handlers.GetHandler{
		GetProfileUC: *usecase.NewGetProfileUseCase(
			postgresClient,
			staticClient,
		),
		GetProfilesUC: *usecase.NewGetProfilesForUserUseCase(
			postgresClient,
			staticClient,
		),
		GetProfileImage: *usecase.NewGetUserPhotoUseCase(
			postgresClient,
			staticClient,
		),
	}

	// NewUser... должен возвращать ошибку
	// done
	// создать свой конструктор с теми же ошибками и тд - потом
	sessionHandler := &handlers.SessionHandler{
		LoginUC: *usecase.NewUserLogInUseCase(
			postgresClient,
			redisClient,
			hasher,
			validator,
		),
		CheckSessionUC: *usecase.NewUserCheckSessionUseCase(
			redisClient,
		),
		LogoutUC: *usecase.NewUserLogOutUseCase(
			postgresClient,
			redisClient,
		),
	}

	profileHandler := &handlers.ProfileHandler{
		LikeUC: *usecase.NewProfileLikeCase(
			postgresClient,
		),
		MatchUC: *usecase.NewProfileMatchCase(
			postgresClient,
		),
		UpdateUC: *usecase.NewProfileUpdateUseCase(
			postgresClient,
		),
		GetProfileUC: *usecase.NewGetProfileUseCase(
			postgresClient,
			staticClient,
		),
		GetProfileImageUC: *usecase.NewGetUserPhotoUseCase(
			postgresClient,
			staticClient,
		),
	}

	userHandler := &handlers.UserHandler{
		SignupUC: *usecase.NewUserSignUpUseCase(
			postgresClient,
			staticClient,
			hasher,
			validator,
		),
		DeleteUserUC: *usecase.NewUserDeleteUseCase(
			postgresClient,
		),
	}

	staticHandler := &handlers.StaticHandler{
		UploadUC: *usecase.NewStaticUploadCase(
			postgresClient,
			staticClient,
		),
		DeleteUC: *usecase.NewStaticDeleteCase(
			postgresClient,
			staticClient,
		),
	}

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	r.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	r.HandleFunc("/users/login", sessionHandler.LoginUser).Methods("POST")
	r.HandleFunc("/users/logout", sessionHandler.LogoutUser).Methods("POST")
	r.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")
	r.HandleFunc("/users/checkSession", sessionHandler.CheckSession).Methods("GET")

	r.HandleFunc("/profiles/{id}", getHandler.GetProfile).Methods("GET")
	r.HandleFunc("/profiles", getHandler.GetProfiles).Methods("GET")
	r.HandleFunc("/profiles/like", profileHandler.SetLike).Methods("POST")
	r.HandleFunc("/profiles/match/{id}", profileHandler.GetMatches).Methods("GET")

	r.HandleFunc("/profiles/uploadPhoto", staticHandler.UploadPhoto).Methods("POST")
	r.HandleFunc("/profiles/deletePhoto", staticHandler.DeletePhoto).Methods("DELETE")

	r.HandleFunc("/profiles/update", profileHandler.UpdateProfile).Methods("POST")

	// r.Use(handlery.AdminAuthMiddleware(sessionHandler))
	// r.Use(handlery.PanicMiddleware)

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

// нужно найти нормальный клиент работы с базой данных чтобы смотреть
// как с ней взаимодействует всё
// https://dbeaver.io/
