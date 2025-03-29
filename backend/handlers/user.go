package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

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

	for _, existingUser := range Users {
		if existingUser.Login == input.Login {
			makeResponse(w, http.StatusConflict, map[string]string{"message": "user already exists"})
			return
		}
	}

	id := len(Users) + 1
	user := config.User{
		UserId:   id,
		Login:    input.Login,
		Password: utils.EncryptPasswordSHA256(input.Password),
	}

	Users[id] = user

	muProfiles.Lock()
	defer muProfiles.Unlock()

	// profiles[id] = config.Profile{
	// 	FirstName:   input.Login,
	// 	LastName:    "Иванов",
	// 	Description: "lalalalalalalala",
	// 	// Birthday: struct {
	// 	// 	Year  int `yaml:"year" json:"year"`
	// 	// 	Month int `yaml:"month" json:"month"`
	// 	// 	Day   int `yaml:"day" json:"day"`
	// 	// }{
	// 	// 	Year:  2005,
	// 	// 	Month: 3,
	// 	// 	Day:   28,
	// 	// },
	// }

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
