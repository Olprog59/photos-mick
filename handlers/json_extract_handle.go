package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/Olprog59/golog"
	"github.com/Olprog59/photos-mick/models"
)

func JsonExtract(medias []models.MyMedia) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, err := json.MarshalIndent(medias, "", " ")
		if err != nil {
			golog.Err("%+v", err)
		}

		_ = os.WriteFile("photos.json", file, 0644)

		w.Header().Add("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(medias)
		if err != nil {
			golog.Err("%+v", err)
		}
	}
}
