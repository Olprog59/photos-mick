package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Olprog59/golog"
	"github.com/Olprog59/photos-mick/commons"
	"github.com/Olprog59/photos-mick/models"
)

func FindAll(medias []models.MyMedia) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// handleGeneric(w, medias[:30], "media_list")
		handleGeneric(w, medias, "media_list")
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

func UpdateMedia(medias *[]models.MyMedia) http.HandlerFunc {
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

		err = r.ParseForm()
		if err != nil {
			http.Error(w, "Problem parse form", http.StatusNotAcceptable)
		}

		newName := r.FormValue("name")

		err = updateMediaInfo(newName, medias, id, w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if (*medias)[id].TypeMedia == "video" {
			// err := updateMetadataVideo(medias[id])
			// if err != nil {
			// 	golog.Err("%s", err.Error())
			// }
		} else {
			m := *medias
			updateMetadataPhoto(&m[id])
		}
		models.ReadFileToSlice(medias)

		file, err := json.MarshalIndent(medias, "", " ")
		if err != nil {
			golog.Err("%+v", err)
		}

		_ = os.WriteFile("photos.json", file, 0644)

		messageHeader(&w, "%s is updated. The media will be reorganising.", (*medias)[id].FileName)

		w.Header().Set("HX-Trigger", "mediaUpdated")
		handleGeneric(w, (*medias)[id], "media_item")
	}
}

// func updateMetadataVideo(media models.MyMedia) error {
// 	formatDateWithoutTimeZone, err := commons.ConvertISOToExifTime(media.Datetime)
// 	if err != nil {
// 		golog.Err("Error converting to Exif time:", err)
// 		return err
// 	}
//
// 	outputFile := "medias/temp_" + media.FileName
//
// 	gpsLatitude := fmt.Sprintf("%.8f", media.Latitude)
// 	gpsLongitude := fmt.Sprintf("%.8f", media.Longitude)
// 	cmd := exec.Command("ffmpeg",
// 		"-i", media.Path,
// 		"-metadata", "creation_time="+formatDateWithoutTimeZone,
// 		"-metadata", "modify_date="+formatDateWithoutTimeZone,
// 		"-metadata", "date="+formatDateWithoutTimeZone,
// 		"-metadata", "DateTimeOriginal="+formatDateWithoutTimeZone,
// 		"-metadata", "time_zone="+media.TimeZone,
// 		"-metadata", "location="+gpsLatitude+","+gpsLongitude,
// 		"-metadata", "description="+media.Description,
// 		"-c", "copy",
// 		"-y",
// 		outputFile,
// 	)
//
// 	golog.Info(cmd.String())
//
// 	// Exécuter la commande
// 	err = cmd.Run()
// 	if err != nil {
// 		return err
// 	}
//
// 	// Remplacer le fichier original par le nouveau fichier
// 	err = os.Rename(outputFile, media.Path)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func updateMetadataPhoto(media *models.MyMedia) {
	// Update metadata with exiftool
	formatDateWithoutTimeZone, err := commons.ConvertISOToExifTime(media.Datetime)
	if err != nil {
		golog.Err("Error converting to Exif time:", err)
		return
	}
	formatDateWithTimeZone := formatDateWithoutTimeZone + media.TimeZone

	gpsDateTime, err := commons.ConvertISOToGPSTime(formatDateWithTimeZone)
	if err != nil {
		golog.Err("Error converting to GPS time:", err)
		return
	}

	gpsLatitude := fmt.Sprintf("%.8f", media.Latitude)
	gpsLongitude := fmt.Sprintf("%.8f", media.Longitude)

	// Déterminer les références de latitude et de longitude
	var gpsLatitudeRef, gpsLongitudeRef string
	if media.Latitude < 0 {
		gpsLatitudeRef = "S"
		gpsLatitude = gpsLatitude[1:] // Supprimer le signe négatif
	} else {
		gpsLatitudeRef = "N"
	}
	if media.Longitude < 0 {
		gpsLongitudeRef = "W"
		gpsLongitude = gpsLongitude[1:] // Supprimer le signe négatif
	} else {
		gpsLongitudeRef = "E"
	}

	ext := path.Ext(media.Path)
	newName := path.Join(path.Dir(media.Path), media.Name+ext)
	newName = strings.ToLower(newName)

	// Construire la commande exiftool
	cmd := exec.Command("exiftool",
		"-n",
		"-FileModifyDate="+formatDateWithoutTimeZone,
		"-CreateDate="+formatDateWithoutTimeZone,
		"-ModifyDate="+formatDateWithoutTimeZone,
		"-Composite:ModifyDate="+formatDateWithoutTimeZone,
		"-DateTimeOriginal="+formatDateWithoutTimeZone,
		"-ExifIFD:OffsetTime="+media.TimeZone,
		"-ExifIFD:OffsetTimeOriginal="+media.TimeZone,
		"-ExifIFD:OffsetTimeDigitized="+media.TimeZone,
		"-TrackModifyDate="+formatDateWithoutTimeZone,
		"-MediaCreateDate="+formatDateWithoutTimeZone,
		"-MediaModifyDate="+formatDateWithoutTimeZone,
		"-CreationDate="+formatDateWithoutTimeZone,
		"-GPS:GPSLatitude="+gpsLatitude,
		"-GPS:GPSLatitudeRef="+gpsLatitudeRef,
		"-GPS:GPSLongitude="+gpsLongitude,
		"-GPS:GPSLongitudeRef="+gpsLongitudeRef,
		"-GPSLatitude="+gpsLatitude,
		"-GPSLatitudeRef="+gpsLatitudeRef,
		"-GPSLongitude="+gpsLongitude,
		"-GPSLongitudeRef="+gpsLongitudeRef,
		"-GPSDateTime="+gpsDateTime,
		"-GPS:GPSDateTime="+gpsDateTime,
		"-GPS:GPSDateStamp="+gpsDateTime[:10],
		"-Exif:ImageDescription="+media.Description,
		"-Description="+media.Description,
		media.Path,
	)

	// golog.Info(cmd.String())
	golog.Info("%+v", media)

	// Exécuter la commande et capturer l'erreur, le cas échéant
	err = cmd.Run()
	if err != nil {
		golog.Err("Failed to update metadata: %v", err)
	}

	err = os.Rename(media.Path, newName)
	if err != nil {
		golog.Err("%+v", err)
	}
	media.Path = newName
	media.FileName = path.Base(newName)
	golog.Success("Updating metadata for %+v", media)
}

func updateMediaInfo(newName string, medias *[]models.MyMedia, id int, w http.ResponseWriter, r *http.Request) error {
	var err error
	// if newName != medias[id].Name {
	// 	ext := filepath.Ext(medias[id].FileName)
	// 	path := newName + ext
	// 	err = os.Rename("medias/"+medias[id].FileName, "medias/"+path)
	// 	if err != nil {
	// 		log.Println(err)
	// 		http.Error(w, err.Error(), http.StatusBadRequest)
	// 		return err
	// 	}
	// 	medias[id].FileName = path
	// }
	//
	(*medias)[id].Name = newName

	dateFormat := models.TestInputDateTime(r.FormValue("date_time"))

	parsedTime, err := time.Parse(dateFormat, r.FormValue("date_time"))
	if err != nil {
		log.Println("Error parsing date_time:", err)
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return err
	}
	(*medias)[id].Datetime = parsedTime.Format(commons.DateFormatWithSeconds)
	(*medias)[id].TimeZone = r.FormValue("time_zone")
	(*medias)[id].Description = r.FormValue("description")
	(*medias)[id].Longitude = commons.StringToFloat64(r.FormValue("longitude"))
	(*medias)[id].Latitude = commons.StringToFloat64(r.FormValue("latitude"))
	(*medias)[id].Modified = true
	return nil
}

// Fonction utilitaire pour recharger les templates (utile en développement)
func ReloadTemplates() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Templates rechargés"))
	}
}

func CalculateCollectionStats(medias []models.MyMedia) http.HandlerFunc {
	stats := models.CollectionStats{}
	stats.TotalMedias = len(medias)

	var oldestTime, newestTime time.Time

	for _, media := range medias {
		// Compter les photos et les vidéos
		switch media.TypeMedia {
		case "photo":
			stats.PhotoCount++
		case "video":
			stats.VideoCount++
		}

		// Trouver la date la plus ancienne et la plus récente
		mediaTime, err := time.Parse(commons.DateFormatWithSeconds, media.Datetime)
		if err == nil {
			if oldestTime.IsZero() || mediaTime.Before(oldestTime) {
				oldestTime = mediaTime
			}
			if newestTime.IsZero() || mediaTime.After(newestTime) {
				newestTime = mediaTime
			}
		}
	}

	stats.OldestDate = oldestTime.Format("02/01/2006")
	stats.NewestDate = newestTime.Format("02/01/2006")

	return func(w http.ResponseWriter, r *http.Request) {
		handleGeneric(w, stats, "stats_panel")
	}
}
