package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Olprog59/golog"
	"github.com/Olprog59/photos-mick/commons"
	"github.com/evanoberholster/imagemeta"
	"github.com/evanoberholster/imagemeta/exif2"
)

type MyMedia struct {
	Description string `json:"description"`
	FileName    string `json:"original_name"`
	Name        string
	Path        string  `json:"path"`
	TimeZone    string  `json:"time_zone"`
	TypeMedia   string  `json:"type"`
	Datetime    string  `json:"date_time"`
	Format      string  `json:"format"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Id          int     `json:"id"`
	Duration    int     `json:"duree"`
	Modified    bool    `json:"modified"`
	Journee     int8    `json:"journee"`
}

func ReadFileToSlice(medias *[]MyMedia) {
	// sortByDatetime(*medias)
	sortByDatetimeAndTimezone(medias)

	for i := 0; i < len(*medias); i++ {
		(*medias)[i].Id = i
	}
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

// func sortByDatetime(medias []MyMedia) {
// 	sort.Slice(medias, func(i, j int) bool {
// 		ti, err1 := time.Parse(commons.DateFormatWithSeconds, medias[i].Datetime)
// 		tj, err2 := time.Parse(commons.DateFormatWithSeconds, medias[j].Datetime)
// 		if err1 != nil || err2 != nil {
// 			return false
// 		}
// 		return ti.Before(tj)
// 	})
// }

//	func writeFile(medias []MyMedia) {
//		jsonData, err := json.MarshalIndent(medias, " ", " ")
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		if _, err := os.Stat("output.json"); os.IsExist(err) {
//			os.Remove("output.json")
//		}
//
//		file, err := os.Create("output.json")
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		defer file.Close()
//
//		_, err = file.Write(jsonData)
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		log.Println("Les données JSON ont été écrites dans le fichier output.json")
//	}

func getMetadata(photo string) (*exif2.Exif, error) {
	file, err := os.Open(photo)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	defer file.Close()

	x, err := imagemeta.Decode(file)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &x, nil
}

// parseGPSCoordinates parses a GPS coordinate string and returns the latitude and longitude.
func parseGPSCoordinates(coord string) (float64, float64, error) {
	// Regular expression to match the GPS coordinate pattern
	re := regexp.MustCompile(`^([+-]\d+\.\d+)([+-]\d+\.\d+)`)

	// Find the matches
	matches := re.FindStringSubmatch(coord)
	if len(matches) != 3 {
		return 0, 0, fmt.Errorf("invalid GPS coordinate format")
	}

	// Parse latitude and longitude
	lat, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, 0, err
	}

	lon, err := strconv.ParseFloat(matches[2], 64)
	if err != nil {
		return 0, 0, err
	}

	return lat, lon, nil
}

func getMetadataVideos(file string) (date, zone string, longitude, latitude float64, width, height, duration int) {
	cmd := exec.Command("exiftool", "-json", "-n", file)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		golog.Err("Failed to execute exiftool for file %s: %v", file, err)
		return
	}

	var exifData []map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &exifData); err != nil {
		golog.Err("Failed to unmarshal exiftool output for file %s: %v", file, err)
		return
	}

	if len(exifData) == 0 {
		golog.Warn("No metadata found for file %s", file)
		return
	}

	metadata := exifData[0]

	// Extract date
	datePriority := []string{"ContentCreateDate", "DateTimeOriginal", "CreateDate", "ModifyDate"}
	for _, dateField := range datePriority {
		if dateStr, ok := metadata[dateField].(string); ok && dateStr != "0000:00:00 00:00:00" {
			parsedDate, err := time.Parse("2006:01:02 15:04:05", dateStr)
			if err == nil {
				date = parsedDate.Format(commons.DateFormatWithSeconds)
				break
			}
		}
	}

	// Extract timezone
	zonePriority := []string{"CustomTimeZone", "OffsetTime", "TimeZone"}
	for _, zoneField := range zonePriority {
		if zoneStr, ok := metadata[zoneField].(string); ok && zoneStr != "" {
			zone = zoneStr
			break
		}
	}

	// Extract GPS coordinates
	if customLatitude, ok := metadata["CustomLatitude"].(float64); ok {
		latitude = customLatitude
	} else if gpsLatitude, ok := metadata["GPSLatitude"].(float64); ok {
		latitude = gpsLatitude
	}

	if customLongitude, ok := metadata["CustomLongitude"].(float64); ok {
		longitude = customLongitude
	} else if gpsLongitude, ok := metadata["GPSLongitude"].(float64); ok {
		longitude = gpsLongitude
	}

	// Extract width and height
	if imageWidth, ok := metadata["ImageWidth"].(float64); ok {
		width = int(imageWidth)
	}
	if imageHeight, ok := metadata["ImageHeight"].(float64); ok {
		height = int(imageHeight)
	}

	// Extract duration
	if durationStr, ok := metadata["Duration"].(string); ok {
		durationFloat, err := strconv.ParseFloat(durationStr, 64)
		if err == nil {
			duration = int(durationFloat)
		}
	}

	if date == "" {
		date = "2021-01-01T00:00:00"
		golog.Warn("No valid date found for file %s", file)
	}

	return date, zone, longitude, latitude, width, height, duration
}

func GetMetadataPhotos(folder string) []MyMedia {
	var medias []MyMedia

	err := filepath.Walk(folder, func(path string, info fs.FileInfo, errr error) error {
		if info.IsDir() {
			return nil
		}

		if errr != nil {
			return errr
		}

		name := strings.ToLower(info.Name())
		ext := filepath.Ext(name)

		dateStart, err := time.Parse("2006-01-02", commons.DateStr)
		if err != nil {
			golog.Err("Error: %+v", err)
		}

		switch ext {
		case ".jpeg", ".jpg", ".heic", "webp", ".png", ".gif":
			exif, err := getMetadata(path)
			if err != nil {
				golog.Err("Error: %+v", err)
			}

			zone, _ := exif.DateTimeOriginal().Zone()
			w := exif.ImageWidth
			h := exif.ImageHeight

			format := ""
			if w > h {
				format = "largeur"
			} else if h > w {
				format = "hauteur"
			} else {
				format = "carre"
			}

			if zone == "" || zone == "UTC" {
				zone = "+00:00"
			}

			m := MyMedia{
				FileName:    name,
				Datetime:    exif.DateTimeOriginal().Format(commons.DateFormatWithSeconds),
				Name:        name[:len(name)-len(ext)],
				Path:        path,
				TimeZone:    zone,
				TypeMedia:   "photo",
				Latitude:    exif.GPS.Latitude(),
				Longitude:   exif.GPS.Longitude(),
				Description: exif.ImageDescription,
				Format:      format,
				Journee:     int8(exif.DateTimeOriginal().Sub(dateStart).Hours() / 24),
			}

			medias = append(medias, m)
		case ".mp4", ".mov":
			dateVideo, zone, longitude, latitude, w, h, duration := getMetadataVideos(path)
			// format, err := time.Parse(dateFormatWithSeconds, dateVideo)
			if err != nil {
				golog.Err("Error: %+v", err)
			}
			format := ""
			if w > h {
				format = "largeur"
			} else if h > w {
				format = "hauteur"
			} else {
				format = "carre"
			}

			datetime, err := time.Parse(commons.DateFormatWithSeconds, dateVideo)
			if err != nil {
				golog.Err("Error: %+v", err)
			}

			if zone == "" || zone == "UTC" {
				zone = "+00:00"
			}

			m := MyMedia{
				FileName:  name,
				Datetime:  dateVideo,
				Name:      name[:len(name)-len(ext)],
				Path:      path,
				TimeZone:  zone,
				TypeMedia: "video",
				Longitude: longitude,
				Latitude:  latitude,
				Format:    format,
				Duration:  duration,
				Journee:   int8(datetime.Sub(dateStart).Hours() / 24),
			}

			medias = append(medias, m)
		}

		return nil
	})
	if err != nil {
		golog.Err("Error: %+v", err)
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
