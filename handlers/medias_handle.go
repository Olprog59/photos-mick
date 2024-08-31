package handlers

import (
	"encoding/json"
	"fmt"
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
		golog.Info("Media found: %v", media)
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
			// err := updateMetadataVideo((*medias)[id])
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

func RenameMedia(medias *[]models.MyMedia) http.HandlerFunc {
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
		// rename file to *.removed
		err = os.Rename(media.Path, media.Path+".removed")
		if err != nil {
			golog.Err("%s", err.Error())
			return
		}

		// delete media from slice
		*medias = append((*medias)[:id], (*medias)[id+1:]...)

		models.ReadFileToSlice(medias)

		messageHeader(&w, "%s a été supprimé. (renommé avec .removed)", media.FileName)
		w.Header().Set("HX-Trigger", "mediaUpdated")
		// handleGeneric(w, media, "media_item")
		fmt.Fprint(w, "Media removed")
	}
}

func updateMetadataVideo(media models.MyMedia) error {
	formatDateWithoutTimeZone, err := commons.ConvertISOToExifTime(media.Datetime)
	if err != nil {
		golog.Err("Error converting to Exif time: %v", err)
		return err
	}

	tempFile := media.Path + ".temp.mp4"

	// Copier le fichier original vers le fichier temporaire
	if err := copyFile(media.Path, tempFile); err != nil {
		golog.Err("Failed to create temporary file: %v", err)
		return err
	}

	// Préparer la commande ExifTool pour une réécriture complète des métadonnées
	cmd := exec.Command("exiftool",
		"-overwrite_original",
		"-P",                     // Préserver la date de modification du fichier
		"-api", "QuickTimeUTC=1", // Traiter les dates QuickTime comme UTC
		fmt.Sprintf("-CreateDate=%s", formatDateWithoutTimeZone),
		fmt.Sprintf("-ModifyDate=%s", formatDateWithoutTimeZone),
		fmt.Sprintf("-MediaCreateDate=%s", formatDateWithoutTimeZone),
		fmt.Sprintf("-MediaModifyDate=%s", formatDateWithoutTimeZone),
		fmt.Sprintf("-TrackCreateDate=%s", formatDateWithoutTimeZone),
		fmt.Sprintf("-TrackModifyDate=%s", formatDateWithoutTimeZone),
		fmt.Sprintf("-DateTimeOriginal=%s", formatDateWithoutTimeZone),
		fmt.Sprintf("-ContentCreateDate=%s", formatDateWithoutTimeZone),
		fmt.Sprintf("-GPSLatitude=%f", media.Latitude),
		fmt.Sprintf("-GPSLongitude=%f", media.Longitude),
		fmt.Sprintf("-GPSLatitudeRef=%s", getLatitudeRef(media.Latitude)),
		fmt.Sprintf("-GPSLongitudeRef=%s", getLongitudeRef(media.Longitude)),
		fmt.Sprintf("-OffsetTime=%s", media.TimeZone),
		fmt.Sprintf("-TimeZone=%s", media.TimeZone),
		fmt.Sprintf("-Description=%s", media.Description),
		// Champs personnalisés
		fmt.Sprintf("-XMP-x:CustomTimeZone=%s", media.TimeZone),
		fmt.Sprintf("-XMP-x:CustomLatitude=%f", media.Latitude),
		fmt.Sprintf("-XMP-x:CustomLongitude=%f", media.Longitude),
		tempFile,
	)

	golog.Debug("Executing ExifTool command: %s", cmd.String())

	output, err := cmd.CombinedOutput()
	if err != nil {
		golog.Err("ExifTool command failed: %v\nOutput: %s", err, string(output))
		os.Remove(tempFile)
		return fmt.Errorf("exiftool failed: %w", err)
	}

	// Vérifier si le fichier temporaire est lisible
	if !isVideoPlayable(tempFile) {
		golog.Err("Temporary file is not playable after metadata update")
		os.Remove(tempFile)
		return fmt.Errorf("temporary file became unplayable after metadata update")
	}

	// Remplacer l'ancien fichier par le nouveau
	if err := os.Rename(tempFile, media.Path); err != nil {
		golog.Err("Failed to replace original file: %v", err)
		os.Remove(tempFile)
		return err
	}

	golog.Info("Video metadata updated successfully for file: %s", media.Path)
	return nil
}

func getLatitudeRef(latitude float64) string {
	if latitude >= 0 {
		return "N"
	}
	return "S"
}

func getLongitudeRef(longitude float64) string {
	if longitude >= 0 {
		return "E"
	}
	return "W"
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}

func isVideoPlayable(filePath string) bool {
	cmd := exec.Command("ffprobe", "-v", "error", filePath)
	return cmd.Run() == nil
}

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
		golog.Debug("Error parsing date_time:", err)
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
