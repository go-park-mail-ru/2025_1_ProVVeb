package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/backend/utils"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/config"
	"github.com/gorilla/mux"
)

type UserHandler struct{}

var Users = utils.InitUserMap()

func (u *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "invalid JSON data"})
		return
	}

	if input.Login == "" || input.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "login and password are required"})
		return
	}

	if (utils.ValidateLogin(input.Login) != nil) || (utils.ValidatePassword(input.Password) != nil) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "invalid email or password"})
		return
	}

	for _, existingUser := range Users {
		if existingUser.Login == input.Login {
			w.WriteHeader(http.StatusConflict)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"message": "user already exists"})
			return
		}
	}

	id := len(Users) + 1
	user := config.User{
		Id:       id,
		Login:    input.Login,
		Password: utils.EncryptPasswordSHA256(input.Password),
	}

	Users[id] = user
	profiles[id] = config.Profile{
		FirstName:   input.Login,
		LastName:    "Иванов",
		Description: "lalalalalalalala",
		Birthday: struct {
			Year  int `yaml:"year" json:"year"`
			Month int `yaml:"month" json:"month"`
			Day   int `yaml:"day" json:"day"`
		}{
			Year:  2005,
			Month: 3,
			Day:   28,
		},
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "user created"})
}

func (u *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	userId, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid user ID"})
		return
	}

	if _, exists := Users[userId]; !exists {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "User not found"})
		return
	}

	delete(Users, userId)
	delete(profiles, userId)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("User with ID %d deleted", userId),
	})
}
