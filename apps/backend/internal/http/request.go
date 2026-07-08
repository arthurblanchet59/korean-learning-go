package http

import (
	"encoding/json"
	"net/http"
)

type answerRequest struct {
	Rating string `json:"rating"`
}

func readJSON(r *http.Request, payload any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(payload)
}
