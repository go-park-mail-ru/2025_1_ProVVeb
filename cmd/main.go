package main

import (
	"fmt"
	"net/http"
	"time"

	handlery "github.com/go-park-mail-ru/2025_1_ProVVeb/delivery"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/repository"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/usecase"

	"github.com/gorilla/mux"

	"github.com/rs/cors"
)

func main() {
	redisAddr := "redis:6379"
	redisDB := 0

	postgresClient, err := repository.NewUserRepo()
	if err != nil {
		fmt.Println(fmt.Errorf("Not able to work with postgresClient: %v", err))
		return
	}
	defer postgresClient.CloseRepo()

	redisClient, err := repository.NewSessionRepo(redisAddr, redisDB)
	if err != nil {
		fmt.Println(fmt.Errorf("Not able to work with redisClient: %v", err))
		return
	}
	defer redisClient.CloseRepo()

	hasher, err := repository.NewPassHasher()
	if err != nil {
		fmt.Println(fmt.Errorf("Not able to work with hasher: %v", err))
		return
	}

	r := mux.NewRouter()

	// getHandler := &handlery.GetHandler{DB: conn}

	// NewUser... должен возвращать ошибку
	// done
	// создать свой конструктор с теми же ошибками и тд - потом
	sessionHandler := &handlery.SessionHandler{
		LoginUC: *usecase.NewUserLogInUseCase(
			postgresClient,
			redisClient,
			hasher,
		),
	}

	// userHandler := &handlery.UserHandler{DB: conn}

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// r.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	r.HandleFunc("/users/login", sessionHandler.LoginUser).Methods("POST")
	// r.HandleFunc("/users/logout", sessionHandler.LogoutUser).Methods("POST")
	// r.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")
	// r.HandleFunc("/users/checkSession", sessionHandler.CheckSession).Methods("GET")

	// r.HandleFunc("/profiles/{id}", getHandler.GetProfile).Methods("GET")
	// r.HandleFunc("/profiles", getHandler.GetProfiles).Methods("GET")

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
	server.ListenAndServe()
}

// нужно найти нормальный клиент работы с базой данных чтобы смотреть
// как с ней взаимодействует всё
// https://dbeaver.io/
