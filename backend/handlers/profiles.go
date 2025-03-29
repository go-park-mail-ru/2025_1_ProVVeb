package handlers

import (
	"net/http"
	"strconv"
	"sync"

	postgres "github.com/go-park-mail-ru/2025_1_ProVVeb/backend/database_function/postgres/queries"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/config"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

const for_single_profile = 30

var muProfiles = &sync.Mutex{}

type GetHandler struct {
	DB *pgx.Conn
}

func (p *GetHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	profileID, err := strconv.Atoi(id)
	if err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid user ID"})
		return
	}

	var profile config.Profile
	profile, err = postgres.DBGetProfilePostgres(p.DB, profileID)

	if err != nil {
		makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "Error fetching profile"})
		return
	}

	makeResponse(w, http.StatusOK, profile)
}

func (p *GetHandler) GetProfiles(w http.ResponseWriter, r *http.Request) {
	var userId string = r.URL.Query().Get("forUser")

	profileId, err := strconv.Atoi(userId)
	if err != nil {
		makeResponse(w, http.StatusBadRequest, map[string]string{"message": "Invalid user ID"})
		return
	}
	muProfiles.Lock()
	defer muProfiles.Unlock()

	profileList := make([]config.Profile, 0, for_single_profile)

	for i := range for_single_profile {
		if i != profileId {
			var profile config.Profile
			profile, err = postgres.DBGetProfilePostgres(p.DB, i)

			if err != nil {
				makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "Error fetching profile"})
				return
			}

			profileList = append(profileList, profile)
		}

	}

	makeResponse(w, http.StatusOK, profileList)
}
