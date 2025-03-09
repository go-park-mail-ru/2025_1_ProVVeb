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
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	profile, exists := profiles[profileID]
	if !exists {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profile)
}

func (p *GetHandler) GetProfiles(w http.ResponseWriter, r *http.Request) {
	var userId string = r.URL.Query().Get("forUser")

	profileId, err := strconv.Atoi(userId)
	if err != nil {

		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	profileList := make([]config.Profile, 0, len(profiles))
	for i, profile := range profiles {
		if i != profileId {
			profileList = append(profileList, profile)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profileList)
}
