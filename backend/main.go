package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/backend/db/postgres"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/backend/db/redis"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/backend/handlers"
	"github.com/gorilla/mux"

	"github.com/rs/cors"
)

func main() {
	redisAddr := "127.0.0.1:6379"
	redisDB := 0

	redisClient := redis.DBInitRedisConfig(redisAddr, redisDB)
	defer redisClient.Close()

	cfg := postgres.DBInitPostgresConfig()

	conn, err := postgres.DBInitConnectionPostgres(cfg)
	if err != nil {
		log.Fatal("Не удалось подключиться к базе данных:", err)
	}
	defer postgres.DBCloseConnectionPostgres(conn)

	r := mux.NewRouter()

	getHandler := &handlers.GetHandler{DB: conn}
	sessionHandler := &handlers.SessionHandler{DB: conn, RedisClient: redisClient}
	userHandler := &handlers.UserHandler{DB: conn}

	// r.Use(handlers.AdminAuthMiddleware(sessionHandler))
	r.Use(handlers.PanicMiddleware)

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	r.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	r.HandleFunc("/users/login", sessionHandler.LoginUser).Methods("POST")
	r.HandleFunc("/users/logout", sessionHandler.LogoutUser).Methods("POST")
	r.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")
	r.HandleFunc("/users/checkSession", sessionHandler.CheckSession).Methods("GET")

	r.HandleFunc("/profiles/{id}", getHandler.GetProfile).Methods("GET")
	r.HandleFunc("/profiles", getHandler.GetProfiles).Methods("GET")

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
