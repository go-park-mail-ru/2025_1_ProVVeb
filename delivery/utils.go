package handlers

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/go-park-mail-ru/2025_1_ProVVeb/model"
	"github.com/mailru/easyjson"
	"github.com/mailru/easyjson/jwriter"
)

func MakeEasyJSONResponse(w http.ResponseWriter, statusCode int, v easyjson.Marshaler) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	data, err := easyjson.Marshal(v)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func MakeUserResponse(w http.ResponseWriter, code int, user model.User) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	var wj jwriter.Writer
	user.MarshalEasyJSON(&wj)
	if wj.Error != nil {
		http.Error(w, wj.Error.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(wj.Buffer.BuildBytes())
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
