package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/backend/handlers"
	"github.com/gorilla/mux"

	"github.com/jackc/pgx/v5"
	"github.com/rs/cors"
)

func main() {
	conn, err := pgx.Connect(context.Background(), "postgresql://postgres:Grey31415@localhost:5432/dev")
	if err != nil {
		log.Fatal("Ошибка подключения к базе данных:", err)
	}
	defer conn.Close(context.Background())

	var message string
	err = conn.QueryRow(context.Background(), "SELECT lastname FROM profiles LIMIT 1").Scan(&message)
	if err != nil {
		log.Fatal("Ошибка при выполнении запроса:", err)
	}
	fmt.Println("Сообщение из базы данных:", message)

	r := mux.NewRouter()

	getHandler := &handlers.GetHandler{DB: conn}
	sessionHandler := &handlers.SessionHandler{}
	userHandler := &handlers.UserHandler{}

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	r.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	r.HandleFunc("/users/login", sessionHandler.LoginUser).Methods("POST")
	r.HandleFunc("/users/logout", sessionHandler.LogoutUser).Methods("POST")
	r.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")
	r.HandleFunc("/users/checkSession", sessionHandler.CheckSession).Methods("GET")

	r.HandleFunc("/profiles/{id}", getHandler.GetProfile).Methods("GET")
	r.HandleFunc("/profiles", getHandler.GetProfiles).Methods("GET")

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://213.219.214.83:8000"},
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
