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

var users = utils.InitUserMap()

func (u *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user config.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user); err != nil {
		http.Error(w, "Invalid user data", http.StatusBadRequest)
		return
	}

	if _, exists := users[user.Id]; exists {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	users[user.Id] = user
	profiles[user.Id] = config.Profile{}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
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
	w.Write([]byte(fmt.Sprintf("User with ID %d deleted", userId)))
}
