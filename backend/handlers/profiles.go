package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/backend/utils"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/config"
	"github.com/gorilla/mux"
)

type GetHandler struct{}

var profiles = utils.InitProfileMap()

func (p *GetHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	profileID, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid user ID"})
		return
	}

	profile, exists := profiles[profileID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Profile not found"})
		return
	}

	profile.Avatar = "http://213.219.214.83:8080/static/" + profile.Avatar
	profile.Card = "http://213.219.214.83:8080/static/" + profile.Card

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profile)
}

func (p *GetHandler) GetProfiles(w http.ResponseWriter, r *http.Request) {
	var userId string = r.URL.Query().Get("forUser")

	profileId, err := strconv.Atoi(userId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid user ID"})
		return
	}
	profileList := make([]config.Profile, 0, len(profiles))
	for i, profile := range profiles {
		if i != profileId {
			profile.Avatar = "http://213.219.214.83:8080/static/" + profile.Avatar
			profile.Card = "http://213.219.214.83:8080/static/" + profile.Card
			profileList = append(profileList, profile)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profileList)
}
