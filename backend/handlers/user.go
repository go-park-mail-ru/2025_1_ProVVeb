package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	postgres "github.com/go-park-mail-ru/2025_1_ProVVeb/backend/db/postgres/queries"
	config "github.com/go-park-mail-ru/2025_1_ProVVeb/backend/objects"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/backend/utils"
	"github.com/gorilla/mux"
)

var muUsers = &sync.Mutex{}

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

	email := fmt.Sprintf("%s@example.com", input.Login)
	phone := fmt.Sprintf("+1234567890%s", input.Login)
	date, _ := time.Parse("2006-01-02", "1990-01-01")

	profile := config.Profile{
		FirstName:   input.Login,
		LastName:    "Иванов",
		IsMale:      true,
		Birthday:    date,
		Height:      180,
		Description: "Do you love communism?",
	}

	user := config.User{
		Login:    input.Login,
		Password: utils.EncryptPasswordSHA256(input.Password),
		Email:    email,
		Phone:    phone,
		Status:   0,
	}

	muProfiles.Lock()
	defer muProfiles.Unlock()

	_, _, err = postgres.DBCreateUserWithProfilePostgres(u.DB, profile, user)
	fmt.Println(err)
	if err != nil {
		makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "Unable to create user and profile"})
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

	err = postgres.DBDeleteUserWithProfilePostgres(u.DB, userId)
	fmt.Println(err)
	if err != nil {
		makeResponse(w, http.StatusNotFound, map[string]string{"message": "Error while deleting user"})
		return
	}

	makeResponse(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("User with ID %d deleted", userId)})
}
