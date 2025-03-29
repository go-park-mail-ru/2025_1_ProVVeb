package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	postgres "github.com/go-park-mail-ru/2025_1_ProVVeb/backend/database_function/postgres/queries"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/backend/utils"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/config"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

var muUsers = &sync.Mutex{}

type UserHandler struct {
	DB *pgx.Conn
}

var Users = utils.InitUserMap()

func (u *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&input); err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "invalid JSON data"})
		return
	}

	if input.Login == "" || input.Password == "" {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "login and password are required"})
		return
	}

	if (utils.ValidateLogin(input.Login) != nil) || (utils.ValidatePassword(input.Password) != nil) {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "invalid email or password"})
		return
	}

	muUsers.Lock()
	defer muUsers.Unlock()

	_, err := postgres.DBGetUserPostgres(u.DB, input.Login)

	if err == nil {
		makeResponse(w, http.StatusNotFound, map[string]string{"message": "User already exists"})
		return
	}

	id := len(Users) + 1
	user := config.User{
		UserId:   id,
		Login:    input.Login,
		Password: utils.EncryptPasswordSHA256(input.Password),
		Email:    "",
		Phone:    "",
		Status:   0,
	}

	_, err = postgres.DBCreateUserPostgres(u.DB, user)
	if err != nil {
		makeResponse(w, http.StatusNotFound, map[string]string{"message": "Unable to create user"})
		return
	}

	muProfiles.Lock()
	defer muProfiles.Unlock()

	date, _ := time.Parse("2006-01-02", "1990-01-01")

	profile := config.Profile{
		FirstName:   input.Login,
		LastName:    "Иванов",
		IsMale:      true,
		Birthday:    date,
		Height:      180,
		Description: "Do you love communism?",
	}

	_, err = postgres.DBCreateProfilePostgres(u.DB, profile)
	if err != nil {
		makeResponse(w, http.StatusNotFound, map[string]string{"message": "Unable to create profile"})
		return
	}

	makeResponse(w, http.StatusCreated, map[string]string{"message": "user created"})
}

func (u *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	userId, err := strconv.Atoi(id)
	if err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid user ID"})
		return
	}

	muUsers.Lock()
	defer muUsers.Unlock()

	if _, exists := Users[userId]; !exists {
		makeResponse(w, http.StatusNotFound, map[string]string{"message": "User not found"})
		return
	}

	delete(Users, userId)

	muProfiles.Lock()
	defer muProfiles.Unlock()

	// delete(profiles, userId)

	makeResponse(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("User with ID %d deleted", userId)})
}
