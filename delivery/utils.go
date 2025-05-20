package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
)

func MakeResponse(w http.ResponseWriter, statusCode int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(v)
}

func GetUrlForMetrics(url string) string {
	pattern := `^\d+$`
	re := regexp.MustCompile(pattern)
	newUrl := strings.Split(url, "/")
	for i, s := range newUrl {
		if re.MatchString(s) {
			newUrl[i] = "{digits}"
		}
	}
	return strings.Join(newUrl, "/")
}
