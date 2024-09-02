package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
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

func extractMedia(medias []models.MyMedia, id string) (*models.MyMedia, int, error) {
	for i, media := range medias {
		if media.Id.String() == id {
			return &media, i, nil
		}
	}
	return nil, -1, fmt.Errorf("media not found")
}

func FindByID(medias *[]models.MyMedia) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		media, _, err := extractMedia(*medias, id)
		if err != nil {
			golog.Err("Media not found: %v", err)
			return
		}

		handleGeneric(w, media, "media_form")
	}
}

func UpdateMedia(medias *[]models.MyMedia) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		media, i, err := extractMedia(*medias, id)
		if err != nil {
			golog.Err("Media not found: %v", err)
			http.Error(w, "Media not found", http.StatusBadRequest)
			return
		}

		err = r.ParseForm()
		if err != nil {
			http.Error(w, "Problem parse form", http.StatusNotAcceptable)
			return
		}

		newName := r.FormValue("name")

		err = updateMediaInfo(newName, &(*medias)[i], r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		media = &(*medias)[i]

		// if media.TypeMedia != "video" {
		updateMetadataPhoto(media)
		// }

		golog.Debug("Media updated: %+v", media)
		models.ReadFileToSlice(medias)

		file, err := json.MarshalIndent(medias, "", " ")
		if err != nil {
			golog.Err("%+v", err)
		}

		_ = os.WriteFile("photos.json", file, 0644)

		messageHeader(&w, "%s is updated. The media will be reorganising.", media.FileName)

		w.Header().Set("HX-Trigger", "mediaUpdated")
		handleGeneric(w, media, "media_item")
	}
}

func RenameMedia(medias *[]models.MyMedia) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		media, _, err := extractMedia(*medias, id)
		if err != nil {
			golog.Err("Media not found: %v", err)
			return
		}

		// rename file to *.removed
		err = os.Rename(media.Path, media.Path+".removed")
		if err != nil {
			golog.Err("%s", err.Error())
			return
		}

		// delete media from slice and save to file
		removeMedia(*medias, id)

		models.ReadFileToSlice(medias)

		messageHeader(&w, "%s a été supprimé. (renommé avec .removed)", media.FileName)
		w.Header().Set("HX-Trigger", "mediaUpdated")
		// handleGeneric(w, media, "media_item")
		fmt.Fprint(w, "Media removed")
	}
}

func removeMedia(medias []models.MyMedia, id string) ([]models.MyMedia, error) {
	for i, media := range medias {
		if media.Id.String() == id {
			// Supprime l'élément en combinant les parties avant et après
			return append(medias[:i], medias[i+1:]...), nil
		}
	}
	return medias, fmt.Errorf("media with ID %s not found", id)
}

func updateMetadataPhoto(media *models.MyMedia) {
	// Update metadata with exiftool
	formatDateWithoutTimeZone, err := commons.ConvertISOToExifTime(media.Datetime)
	if err != nil {
		golog.Err("Error converting to Exif time:", err)
		return
	}
	formatDateWithTimeZone := formatDateWithoutTimeZone + media.TimeZone
	golog.Info("formatDateWithTimeZone: %s", formatDateWithTimeZone)

	gpsDateTime, err := commons.ConvertISOToGPSTime(formatDateWithTimeZone)
	if err != nil {
		golog.Err("Error converting to GPS time:", err)
		return
	}

	var gpsLatitudeRef, gpsLongitudeRef string
	var gpsLatitude, gpsLongitude string

	// Latitude
	if media.Latitude < 0 {
		gpsLatitudeRef = "S"
	} else {
		gpsLatitudeRef = "N"
	}
	gpsLatitude = fmt.Sprintf("%.6f", media.Latitude)

	// Longitude
	if media.Longitude < 0 {
		gpsLongitudeRef = "W"
	} else {
		gpsLongitudeRef = "E"
	}
	gpsLongitude = fmt.Sprintf("%.6f", media.Longitude)

	ext := path.Ext(media.Path)
	newName := path.Join(path.Dir(media.Path), media.Name+ext)
	newName = strings.ToLower(newName)

	// []string{"ContentCreateDate", "DateTimeOriginal", "CreateDate", "ModifyDate"}

	// Construire la commande exiftool
	cmd := exec.Command("exiftool",
		"-v",
		"-n",
		"-P",
		"-api", "MWG",
		"-overwrite_original",
		"-ContentCreateDate="+formatDateWithoutTimeZone,
		"-FileModifyDate="+formatDateWithoutTimeZone,
		"-CreateDate="+formatDateWithoutTimeZone,
		"-CreationDate="+formatDateWithoutTimeZone,
		"-GPSDateTime="+formatDateWithoutTimeZone+"Z",
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
		"-SubSecCreateDate="+formatDateWithTimeZone,
		"-SubSecDateTimeOriginal="+formatDateWithTimeZone,
		"-SubSecModifyDate="+formatDateWithTimeZone,
		"-QuickTime:CreateDate="+formatDateWithoutTimeZone,
		"-QuickTime:ModifyDate="+formatDateWithoutTimeZone,
		"-QuickTime:TrackCreateDate="+formatDateWithoutTimeZone,
		"-QuickTime:TrackModifyDate="+formatDateWithoutTimeZone,
		"-QuickTime:MediaCreateDate="+formatDateWithoutTimeZone,
		"-QuickTime:MediaModifyDate="+formatDateWithoutTimeZone,
		"-QuickTime:GPSCoordinates="+gpsLatitude+gpsLatitudeRef+", "+gpsLongitude+gpsLongitudeRef,
		"-QuickTime:TimeZone="+media.TimeZone,
		"-QuickTime:UTC+Offset="+media.TimeZone,
		"-XMP:OffsetTimeOriginal="+media.TimeZone,
		"-XMP:OffsetTimeDigitized="+media.TimeZone,
		"-QuickTime:Description="+media.Description,
		"-XMP:Description="+media.Description,
		"-UserData:Description="+media.Description,
		"-QuickTime:TimeZone="+media.TimeZone,
		"-QuickTime:CreationDate="+formatDateWithTimeZone,
		"-QuickTime:CreationDate-fra-FR="+formatDateWithTimeZone,
		"-XMP:OffsetTimeOriginal="+media.TimeZone,
		"-XMP:OffsetTime="+media.TimeZone,
		media.Path,
	)

	golog.Debug("Command: %v", cmd.String())

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

func updateMediaInfo(newName string, media *models.MyMedia, r *http.Request) error {
	var err error
	media.Name = newName

	dateFormat := models.TestInputDateTime(r.FormValue("date_time"))

	parsedTime, err := time.Parse(dateFormat, r.FormValue("date_time"))
	if err != nil {
		return err
	}
	media.Datetime = parsedTime.Format(commons.DateFormatWithSeconds)
	media.TimeZone = r.FormValue("time_zone")
	media.Description = r.FormValue("description")
	media.Longitude = commons.StringToFloat64(r.FormValue("longitude"))
	media.Latitude = commons.StringToFloat64(r.FormValue("latitude"))
	media.Modified = true
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
