package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Olprog59/golog"
	"github.com/Olprog59/photos-mick/commons"
	"github.com/Olprog59/photos-mick/exiftool"
	"github.com/google/uuid"
)

type MyMedia struct {
	Description string `json:"description"`
	FileName    string `json:"original_name"`
	Name        string
	Path        string    `json:"path"`
	TimeZone    string    `json:"time_zone"`
	TypeMedia   string    `json:"type"`
	Datetime    string    `json:"date_time"`
	Format      string    `json:"format"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Id          uuid.UUID `json:"id"`
	Duration    int       `json:"duree"`
	Modified    bool      `json:"modified"`
	Journee     int8      `json:"journee"`
}

func ReadFileToSlice(medias *[]MyMedia) {
	// sortByDatetime(*medias)
	sortByDatetimeAndTimezone(medias)
}

func sortByDatetimeAndTimezone(medias *[]MyMedia) {
	sort.Slice(*medias, func(i, j int) bool {
		if (*medias)[i].TimeZone == "" || (*medias)[i].TimeZone == "UTC" {
			(*medias)[i].TimeZone = "+00:00"
		}

		if (*medias)[j].TimeZone == "" || (*medias)[j].TimeZone == "UTC" {
			(*medias)[j].TimeZone = "+00:00"
		}

		// Concaténer la date et le fuseau horaire
		timeStrI := (*medias)[i].Datetime + (*medias)[i].TimeZone
		timeStrJ := (*medias)[j].Datetime + (*medias)[j].TimeZone

		// golog.Info("dateTimeI: %s", timeStrI)
		// golog.Info("dateTimeJ: %s", timeStrJ)

		// Analyser la chaîne de date avec fuseau horaire
		ti, err1 := time.Parse("2006-01-02T15:04:05-07:00", timeStrI)
		tj, err2 := time.Parse("2006-01-02T15:04:05-07:00", timeStrJ)

		if err1 != nil || err2 != nil {
			// Gérer les erreurs si nécessaire
			return false
		}

		// Comparer les temps en UTC
		return ti.Before(tj)
	})
}

func getMetadata(metadata map[string]any) (*exiftool.Metadata, error) {
	var data exiftool.Metadata

	if path, ok := metadata["SourceFile"].(string); ok {
		data.Path = path
	} else {
		return nil, errors.New("no path found")
	}

	if name, ok := metadata["FileName"].(string); ok {
		data.Name = name
	} else {
		return nil, errors.New("no name found")
	}

	// Extract date
	datePriority := []string{"ContentCreateDate", "DateTimeOriginal", "CreateDate", "ModifyDate"}
	for _, dateField := range datePriority {
		if dateStr, ok := metadata[dateField].(string); ok && dateStr != "0000:00:00 00:00:00" {
			parsedDate, err := time.Parse("2006:01:02 15:04:05", dateStr)
			if err == nil {
				data.Date = parsedDate
				break
			}
		}
	}

	if zone, ok := metadata["CreationDate-fra-FR"].(string); ok {
		golog.Info("zone: %s", zone)
		parsedDate, err := time.Parse("2006:01:02 15:04:05-07:00", zone)
		if err != nil {
			golog.Err("Error: %+v", err)
		}
		data.Zone = parsedDate.Format("-07:00")

	}

	// Extract timezone
	zonePriority := []string{"CustomTimeZone", "OffsetTime", "TimeZone"}
	for _, zoneField := range zonePriority {
		if zoneStr, ok := metadata[zoneField].(string); ok && zoneStr != "" {
			data.Zone = zoneStr
			break
		}
	}

	// Extract GPS coordinates
	if customLatitude, ok := metadata["CustomLatitude"].(float64); ok {
		data.Lat = customLatitude
	} else if gpsLatitude, ok := metadata["GPSLatitude"].(float64); ok {
		data.Lat = gpsLatitude
	}

	if customLongitude, ok := metadata["CustomLongitude"].(float64); ok {
		data.Lon = customLongitude
	} else if gpsLongitude, ok := metadata["GPSLongitude"].(float64); ok {
		data.Lon = gpsLongitude
	}

	// Extract width and height
	if imageWidth, ok := metadata["ImageWidth"].(float64); ok {
		data.Width = int(imageWidth)
	}
	if imageHeight, ok := metadata["ImageHeight"].(float64); ok {
		data.Height = int(imageHeight)
	}

	// Extract duration
	if durationStr, ok := metadata["Duration"].(string); ok {
		durationFloat, err := strconv.ParseFloat(durationStr, 64)
		if err == nil {
			data.Duration = int(durationFloat)
		}
	}
	descriptions := []string{"ImageDescription", "Description"}
	for _, description := range descriptions {
		if desc, ok := metadata[description].(string); ok && desc != "" {
			data.Description = desc
			break
		}
	}

	if data.Date.IsZero() {
		data.Date = time.Time{}
		golog.Warn("No valid date found for file %s", data.Path)
	}

	return &data, nil
}

func GetMetadataAllInFolder(folder string) (*[]*exiftool.Metadata, error) {
	folder = filepath.Clean(folder)
	golog.Info("folder: %s", folder)
	cmd := exec.Command("exiftool", "-json", "-n", folder)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		golog.Err("Failed to execute exiftool for file %s: %v", folder, err)
		golog.Err("cmd: %v", cmd)
		return nil, err
	}

	var exifData []map[string]any
	if err := json.Unmarshal(out.Bytes(), &exifData); err != nil {
		golog.Err("Failed to unmarshal exiftool output for file %s: %v", folder, err)
		return nil, err
	}

	if len(exifData) == 0 {
		golog.Warn("No metadata found for file %s", folder)
		return nil, errors.New("no metadata found")
	}

	var datas []*exiftool.Metadata
	for _, metadata := range exifData {
		data, err := getMetadata(metadata)
		if err != nil {
			golog.Err("Error: %+v", err)
			continue
		}
		datas = append(datas, data)

	}

	return &datas, nil
}

func GetMetadataAll(folder string) []MyMedia {
	var medias []MyMedia

	m, err := GetMetadataAllInFolder(folder)
	if err != nil {
		golog.Err("Error: %+v", err)
		return nil
	}

	for _, metadata := range *m {

		name := strings.ToLower(metadata.Name)
		ext := filepath.Ext(name)

		dateStart, err := time.Parse("2006-01-02", commons.DateStr)
		if err != nil {
			golog.Err("Error: %+v", err)
		}

		if ext == ".jpeg" || ext == ".jpg" || ext == ".webp" || ext == ".mov" || ext == ".mp4" {

			w := metadata.Width
			h := metadata.Height

			format := ""
			if w > h {
				format = "largeur"
			} else if h > w {
				format = "hauteur"
			} else {
				format = "carre"
			}

			if metadata.Zone == "" || metadata.Zone == "UTC" {
				metadata.Zone = "+00:00"
			}

			uid, err := uuid.NewV7()
			if err != nil {
				golog.Debug("uuid error: %+v", err)
			}

			m := MyMedia{
				Description: metadata.Description,
				FileName:    metadata.Name,
				TimeZone:    metadata.Zone,
				Format:      format,
				Datetime:    metadata.Date.Format(commons.DateFormatWithSeconds),
				Name:        name[:len(name)-len(ext)],
				Path:        metadata.Path,
				Longitude:   metadata.Lon,
				Latitude:    metadata.Lat,
				Id:          uid,
				Duration:    metadata.Duration,
				Journee:     int8(metadata.Date.Sub(dateStart).Hours() / 24),
			}

			if ext == ".mov" || ext == ".mp4" {
				m.TypeMedia = "video"
			} else {
				m.TypeMedia = "photo"
			}

			medias = append(medias, m)
		}
	}

	return medias
}

func TestInputDateTime(input string) string {
	re := regexp.MustCompile(`^(?P<year>\d{4})-(?P<month>\d{2})-(?P<day>\d{2})T(?P<hour>\d{2}):(?P<minute>\d{2})(:(?P<second>\d{2}))?$`)
	match := re.FindStringSubmatch(input)

	// Create a map for named capturing groups
	groupNames := re.SubexpNames()
	matches := map[string]string{}
	for i, name := range groupNames {
		if i != 0 && name != "" {
			matches[name] = match[i]
		}
	}

	// Check if the "second" group matched a value
	if second, ok := matches["second"]; ok && second != "" {
		return commons.DateFormatWithSeconds
	}

	return commons.DateFormatWithoutSeconds
}
