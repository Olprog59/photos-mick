package models

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"math"
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
	sortByDatetime(*medias)

	for i := 0; i < len(*medias); i++ {
		(*medias)[i].Id = i
	}
}

func sortByDatetime(medias []MyMedia) {
	sort.Slice(medias, func(i, j int) bool {
		ti, err1 := time.Parse(commons.DateFormatWithSeconds, medias[i].Datetime)
		tj, err2 := time.Parse(commons.DateFormatWithSeconds, medias[j].Datetime)
		if err1 != nil || err2 != nil {
			return false
		}
		return ti.Before(tj)
	})
}

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
	cmd := exec.Command("mediainfo", file, "--output=JSON")

	var err error
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	media, err := UnmarshalMediaInfo(out.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	var parsedDate time.Time

	for _, m := range media.Media.Track {
		f, err := strconv.ParseFloat(*m.Duration, 32)
		if err != nil {
			golog.Err("Error: %+v", err)
		}
		duration = int(math.Ceil(f))
	}

	for _, m := range media.Media.Track {
		if m.Width != nil || m.Height != nil {
			width, err = strconv.Atoi(*m.Width)
			if err != nil {
				golog.Err("Error: %+v", err)
			}
			height, err = strconv.Atoi(*m.Height)
			if err != nil {
				golog.Err("Error: %+v", err)
			}
			break
		}
	}

	for _, m := range media.Media.Track {
		if m.Extra != nil {
			if m.Extra.COMAppleQuicktimeCreationdate != nil {
				date = *m.Extra.COMAppleQuicktimeCreationdate
				break
			}
		}
	}

	for _, m := range media.Media.Track {
		if date == "" {
			date = *m.EncodedDate
		}
	}

	for _, m := range media.Media.Track {
		if m.Extra != nil {
			if m.Extra.COMAppleQuicktimeLocationISO6709 != nil {
				latitude, longitude, err = parseGPSCoordinates(*m.Extra.COMAppleQuicktimeLocationISO6709)
				if err != nil {
					log.Println(err)
					break
				}
				break

			}
		}
	}

	// golog.Info("Date: %s", date)
	parsedDate, err = time.Parse("2006-01-02T15:04:05-0700", date)
	if err != nil {
		parsedDate, err = time.Parse("2006-01-02T15:04:05-07:00", date)
		if err != nil {
			parsedDate, err = time.Parse("2006-01-02 15:04:05 UTC", date)
			if err != nil {
				parsedDate, err = time.Parse("2006-01-02T15:04:05", date)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}

	zone, _ = parsedDate.Zone()
	date = parsedDate.Format(commons.DateFormatWithSeconds)
	return date, zone, longitude, latitude, width, height, duration
}

func GetMetadataPhotos(folder string) []MyMedia {
	var medias []MyMedia

	err := filepath.Walk(folder, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		name := strings.ToLower(info.Name())
		ext := filepath.Ext(name)

		dateStart, err := time.Parse("2006-01-02", commons.DateStr)
		if err != nil {
			golog.Err("Error: %+v", err)
		}

		switch ext {
		case ".jpeg", ".jpg", ".heic":
			exif, err := getMetadata(path)
			if err != nil {
				log.Println(err)
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
		log.Println(err)
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
