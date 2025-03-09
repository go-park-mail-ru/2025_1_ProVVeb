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

type UserResponse struct {
	ID string `json:"id"`
}

type UserHandler struct{}

var users = utils.InitUserMap()

func (u *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&input); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	if input.Email == "" || input.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	if (utils.ValidateEmail(input.Email) != nil) || (utils.ValidatePassword(input.Password) != nil) {
		http.Error(w, "Invalid email or password", http.StatusBadRequest)
		return
	}

	for _, existingUser := range users {
		if existingUser.Email == input.Email {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
	}

	id := len(users) + 1
	user := config.User{
		Id:       id,
		Email:    input.Email,
		Password: utils.EncryptPasswordSHA256(input.Password),
	}

	users[id] = user
	profiles[id] = config.Profile{}
	w.WriteHeader(http.StatusCreated)

	response := UserResponse{
		ID: fmt.Sprint(user.Id),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func (u *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	userId, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if _, exists := users[userId]; !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	delete(users, userId)
	delete(profiles, userId)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("User with ID %d deleted", userId),
	})
}
