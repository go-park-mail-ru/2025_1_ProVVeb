package handlers

import (
	"net/http"
	"strconv"
	"sync"

	postgres "github.com/go-park-mail-ru/2025_1_ProVVeb/backend/database_function/postgres/queries"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/backend/utils"
	"github.com/go-park-mail-ru/2025_1_ProVVeb/config"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

var muProfiles = &sync.Mutex{}

type GetHandler struct {
	DB *pgx.Conn
}

var profiles = utils.InitProfileMap()

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

	profileList := make([]config.Profile, 0, len(profiles))
	for i, profile := range profiles {
		if i != profileId {
			profile.Avatar = "http://213.219.214.83:8080/static/" + profile.Avatar
			profile.Card = "http://213.219.214.83:8080/static/" + profile.Card
			profileList = append(profileList, profile)
		}
	}

	makeResponse(w, http.StatusOK, profileList)
}
