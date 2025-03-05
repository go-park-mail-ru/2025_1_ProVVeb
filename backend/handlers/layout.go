package handlers

import (
	"net/http"
)

type LayoutHandler struct{}

var BeginForm = []byte(`
<html>
	<body>
	<h1>А ты записался!</h1>
</html>
`)

func (u *LayoutHandler) MainPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Write(BeginForm)
		return
	}
}
