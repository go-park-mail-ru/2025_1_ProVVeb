package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/backend/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	r := mux.NewRouter()

	getHandler := &handlers.GetHandler{}
	sessionHandler := &handlers.SessionHandler{}
	userHandler := &handlers.UserHandler{}

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	r.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	r.HandleFunc("/users/login", sessionHandler.LoginUser).Methods("POST")
	r.HandleFunc("/users/logout", sessionHandler.LogoutUser).Methods("POST")
	r.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")

	r.HandleFunc("/profiles/{id}", getHandler.GetProfile).Methods("GET")
	r.HandleFunc("/profiles", getHandler.GetProfiles).Methods("GET")

	// r.HandleFunc("/backend/static/{file}", getHandler.GetStaticFile).Methods("GET")

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
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
