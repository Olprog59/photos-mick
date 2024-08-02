package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/evanoberholster/imagemeta"
	"github.com/evanoberholster/imagemeta/exif2"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

type MyMedia struct {
	FileName    string `json:"original_name"`
	Datetime    string `json:"date_time"`
	Name        string
	Path        string  `json:"path"`
	TimeZone    string  `json:"time_zone"`
	TypeMedia   string  `json:"type"`
	Description string  `json:"description"`
	Id          int     `json:"id"`
	Longitude   float64 `json:"longitude"`
	Latitude    float64 `json:"latitude"`
	Modified    bool    `json:"modified"`
}

const (
	dateFormatWithSeconds    = "2006-01-02T15:04:05"
	dateFormatWithoutSeconds = "2006-01-02T15:04"
)

func main() {
	log.Println("Hello")

	medias := getMetadataPhotos()

	readFileToSlice(&medias)

	// sortByDatetime(medias)
	//
	// writeFile(medias)

	mux := http.NewServeMux()
	// mux.HandleFunc("GET /json", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Content-Type", "application/json")
	// 	json.NewEncoder(w).Encode(medias)
	// })

	mux.HandleFunc("GET /api/medias", func(w http.ResponseWriter, r *http.Request) {
		readFileToSlice(&medias)

		handleGeneric(w, medias, "medias.html", "templates/medias.html", "templates/media.html")
	})

	mux.HandleFunc("GET /api/medias/refresh", func(w http.ResponseWriter, r *http.Request) {
		medias = getMetadataPhotos()
		readFileToSlice(&medias)
		messageHeader(&w, "Updated list medias")
	})

	mux.HandleFunc("GET /media/{id}", func(w http.ResponseWriter, r *http.Request) {
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
		handleGeneric(w, medias[id], "media-form.html", "templates/media-form.html", "templates/media.html")
	})

	mux.HandleFunc("PUT /media/{id}", func(w http.ResponseWriter, r *http.Request) {
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

		dateFormat := testInputDateTime(r.FormValue("date_time"))

		parsedTime, err := time.Parse(dateFormat, r.FormValue("date_time"))
		if err != nil {
			log.Println("Error parsing date_time:", err)
			http.Error(w, "Invalid date format", http.StatusBadRequest)
			return
		}
		medias[id].Datetime = parsedTime.Format(dateFormatWithSeconds)
		medias[id].TimeZone = r.FormValue("time_zone")
		medias[id].Description = r.FormValue("description")
		medias[id].Longitude = stringToFloat64(r.FormValue("longitude"))
		medias[id].Latitude = stringToFloat64(r.FormValue("latitude"))
		medias[id].Modified = true

		readFileToSlice(&medias)

		messageHeader(&w, "%s is updated. The media will be reorganising.", medias[id].FileName)

		w.Header().Set("HX-Trigger", "mediaUpdated")
		handleGeneric(w, medias[id], "media.html", "templates/media.html")
	})

	mux.Handle("/", http.FileServer(http.Dir("./static")))
	mux.Handle("/medias/", http.StripPrefix("/medias", http.FileServer(http.Dir("./medias"))))

	log.Println("Started server http://localhost:8080")
	http.ListenAndServe("localhost:8080", mux)
}

func messageHeader(w *http.ResponseWriter, message string, data ...any) {
	(*w).Header().Set("Message", fmt.Sprintf(message+"\n", data...))
	(*w).Header().Set("Content-Type", "text/html; charset=utf-8")
}

func handleGeneric(w http.ResponseWriter, data any, namePage string, htmlPage ...string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	funcs := template.FuncMap{
		"dateFR": func(datetime string) string {
			t, err := time.Parse(dateFormatWithSeconds, datetime)
			if err != nil {
				log.Println(err)
				http.Error(w, "datetime format error", http.StatusBadRequest)
			}
			return t.Format("Monday 02 January 2006 15:04:05")
		},
	}

	tmpl, err := template.New(namePage).Funcs(funcs).ParseFiles(htmlPage...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func testInputDateTime(input string) string {
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
		return dateFormatWithSeconds
	}

	return dateFormatWithoutSeconds
}

func readFileToSlice(medias *[]MyMedia) {
	sortByDatetime(*medias)

	for i := 0; i < len(*medias); i++ {
		(*medias)[i].Id = i
	}
}

func stringToFloat64(str string) float64 {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		log.Println(err)
		return 0
	}
	return f
}

func getMetadataVideos(file string) (date, zone string, longitude, latitude float64) {
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

	parsedDate, err = time.Parse("2006-01-02T15:04:05-0700", date)
	if err != nil {
		parsedDate, err = time.Parse("2006-01-02T15:04:05-07:00", date)
		if err != nil {
			parsedDate, err = time.Parse("2006-01-02 15:04:05 UTC", date)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	zone, _ = parsedDate.Zone()
	date = parsedDate.Format(dateFormatWithSeconds)
	return date, zone, longitude, latitude
}

func getMetadataPhotos() []MyMedia {
	var medias []MyMedia

	filepath.Walk("medias", func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		name := strings.ToLower(info.Name())
		ext := filepath.Ext(name)

		switch ext {
		case ".jpeg", ".jpg", ".heic":
			exif := getMetadata(path)
			zone, _ := exif.DateTimeOriginal().Zone()
			m := MyMedia{
				FileName:  name,
				Datetime:  exif.DateTimeOriginal().Format(dateFormatWithSeconds),
				Name:      name[:len(name)-len(ext)],
				Path:      path,
				TimeZone:  zone,
				TypeMedia: "photo",
				Latitude:  exif.GPS.Latitude(),
				Longitude: exif.GPS.Longitude(),
			}

			medias = append(medias, m)
		case ".mp4", ".mov":
			dateVideo, zone, longitude, latitude := getMetadataVideos(path)
			// format, err := time.Parse(dateFormatWithSeconds, dateVideo)
			if err != nil {
				log.Println(err)
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
			}

			medias = append(medias, m)
		}

		return nil
	})

	return medias
}

// Fonction pour trier les médias par date de création
func sortByDatetime(medias []MyMedia) {
	sort.Slice(medias, func(i, j int) bool {
		ti, err1 := time.Parse(dateFormatWithSeconds, medias[i].Datetime)
		tj, err2 := time.Parse(dateFormatWithSeconds, medias[j].Datetime)
		if err1 != nil || err2 != nil {
			return false
		}
		return ti.Before(tj)
	})
}

func getMetadata(photo string) exif2.Exif {
	file, err := os.Open(photo)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	x, err := imagemeta.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	return x
}

func writeFile(medias []MyMedia) {
	jsonData, err := json.MarshalIndent(medias, " ", " ")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat("output.json"); os.IsExist(err) {
		os.Remove("output.json")
	}

	file, err := os.Create("output.json")
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Les données JSON ont été écrites dans le fichier output.json")
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
