package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/backend/handlers"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	LayoutHandler := handlers.LayoutHandler{}
	getHandler := &handlers.GetHandler{}
	sessionHandler := &handlers.SessionHandler{}
	userHandler := &handlers.UserHandler{}

	r.HandleFunc("/main", LayoutHandler.MainPage)

	r.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	r.HandleFunc("/users/login", sessionHandler.LoginUser).Methods("POST")
	r.HandleFunc("/users/logout", sessionHandler.LogoutUser).Methods("POST")
	r.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")

	r.HandleFunc("/profiles/{id}", getHandler.GetProfile).Methods("GET")
	r.HandleFunc("/profiles", getHandler.GetProfiles).Methods("GET")

	server := http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("starting server at :8080")
	server.ListenAndServe()
}
