package handlers

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

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
	var birth time.Time
	err = p.DB.QueryRow(context.Background(),
		`SELECT 
        p.profile_id, p.firstname, p.lastname, p.is_male,
        p.birthday, p.description, l.country, s.path AS avatar
    FROM profiles p
    LEFT JOIN locations l ON p.location_id = l.location_id
	LEFT JOIN static s ON p.photo_id = s.id
    WHERE p.profile_id = $1`, profileID).Scan(
		&profile.ProfileId, &profile.FirstName, &profile.LastName, &profile.IsMale,
		&birth, &profile.Description, &profile.Location, &profile.Avatar,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			makeResponse(w, http.StatusNotFound, map[string]string{"message": "Profile not found"})
		} else {
			makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "1)Database query error"})
		}
		return
	}

	rows, err := p.DB.Query(context.Background(),
		`SELECT i.description
		FROM profile_interests pi
		JOIN interests i ON pi.interest_id = i.interest_id
		WHERE pi.profile_id = $1`, profileID)
	if err != nil {
		makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "2)Database query error"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var interest string
		if err := rows.Scan(&interest); err != nil {
			makeResponse(w, http.StatusInternalServerError, map[string]string{"message": "Error scanning interests"})
			return
		}
		profile.Interests = append(profile.Interests, interest)
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
