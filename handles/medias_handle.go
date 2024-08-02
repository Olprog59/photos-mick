package handles

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Olprog59/photos-mick/commons"
	"github.com/Olprog59/photos-mick/models"
)

func FindAllByPage(medias []models.MyMedia) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}

		itemsPerPage := 20
		start := (page - 1) * itemsPerPage
		end := start + itemsPerPage
		if end > len(medias) {
			end = len(medias)
		}

		if start >= len(medias) {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		paginatedMedias := medias[start:end]

		// Rendre uniquement les nouveaux éléments médias
		handleGeneric(w, paginatedMedias, "media_list")
	}
}

func FindByID(medias *[]models.MyMedia) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid media ID", http.StatusBadRequest)
			return
		}
		if id < 0 || id >= len(*medias) {
			http.Error(w, "Media not found", http.StatusNotFound)
			return
		}
		media := (*medias)[id]
		log.Printf("Media found: %v", media)
		handleGeneric(w, media, "media_form")
	}
}

func UpdateMedia(medias []models.MyMedia) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid media ID", http.StatusBadRequest)
			return
		}
		if id < 0 || id >= len(medias) {
			http.Error(w, "Media not found", http.StatusNotFound)
			return
		}

		err = r.ParseForm()
		if err != nil {
			http.Error(w, "Problem parse form", http.StatusNotAcceptable)
		}

		newName := r.FormValue("name")

		if newName != medias[id].Name {
			ext := filepath.Ext(medias[id].FileName)
			path := newName + ext
			err = os.Rename("medias/"+medias[id].FileName, "medias/"+path)
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			medias[id].FileName = path
		}
		medias[id].Name = newName

		dateFormat := models.TestInputDateTime(r.FormValue("date_time"))

		parsedTime, err := time.Parse(dateFormat, r.FormValue("date_time"))
		if err != nil {
			log.Println("Error parsing date_time:", err)
			http.Error(w, "Invalid date format", http.StatusBadRequest)
			return
		}
		medias[id].Datetime = parsedTime.Format(commons.DateFormatWithSeconds)
		medias[id].TimeZone = r.FormValue("time_zone")
		medias[id].Description = r.FormValue("description")
		medias[id].Longitude = commons.StringToFloat64(r.FormValue("longitude"))
		medias[id].Latitude = commons.StringToFloat64(r.FormValue("latitude"))
		medias[id].Modified = true

		models.ReadFileToSlice(&medias)

		messageHeader(&w, "%s is updated. The media will be reorganising.", medias[id].FileName)

		w.Header().Set("HX-Trigger", "mediaUpdated")
		handleGeneric(w, medias[id], "media_item")
	}
}

// Fonction utilitaire pour recharger les templates (utile en développement)
func ReloadTemplates() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Templates rechargés"))
	}
}
