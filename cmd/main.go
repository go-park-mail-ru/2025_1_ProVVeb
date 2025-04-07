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
	redisAddr := "localhost:6379"
	redisDB := 0

	slonyara := repository.NewUserRepo()
	defer slonyara.CloseRepo()

	redisClient := repository.NewSessionRepo(redisAddr, redisDB)
	defer redisClient.CloseRepo()

	hasher := repository.NewPassHasher()

	r := mux.NewRouter()

	// getHandler := &handlery.GetHandler{DB: conn}
	sessionHandler := &handlery.SessionHandler{
		LoginUC: *usecase.NewUserLogInUseCase(
			slonyara,
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
