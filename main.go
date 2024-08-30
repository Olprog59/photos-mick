package main

import (
	"embed"
	"encoding/json"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/Olprog59/golog"
	"github.com/Olprog59/photos-mick/commons"
	"github.com/Olprog59/photos-mick/handlers"
	"github.com/Olprog59/photos-mick/middleware"
	"github.com/Olprog59/photos-mick/models"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static
var static embed.FS

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	_ = handlers.ReloadTemplates()
	golog.SetLanguage("fr")
	golog.EnableFileNameLogging()
}

func main() {
	var folder string

	flag.StringVar(&commons.DateStr, "date", "", "date")
	flag.StringVar(&folder, "folder", "", "folder")

	flag.Parse()

	golog.Info("date: %s", commons.DateStr)
	golog.Info("folder: %s", folder)

	if folder == "" {
		golog.Err("Folder is empty")
		return
	}

	// vérification si le dossier existe et indiquer ou se trouve l'utilisateur
	if _, err := os.Stat(folder); err != nil {
		golog.Err("Ce dossier n'existe pas: %s", folder)
		here, err := os.Getwd()
		if err != nil {
			golog.Err("%v", err)
			return
		}
		golog.Debug("Tu te trouves ici: %s", here)
		return
	}

	if commons.DateStr == "" {
		golog.Err("Date is empty")
		return
	}

	golog.Notice("Welcome. Started application and analyse medias.")

	medias := models.GetMetadataPhotos(folder)

	models.ReadFileToSlice(&medias)

	handlers.InitHandlers(templateFS)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", handlers.IndexHandle())

	mux.HandleFunc("GET /api/medias/{$}", handlers.FindAll(medias))

	mux.HandleFunc("GET /api/medias/stats", handlers.CalculateCollectionStats(medias))

	mux.HandleFunc("GET /api/media/{id}", handlers.FindByID(&medias))

	mux.HandleFunc("PUT /api/media/{id}", handlers.UpdateMedia(&medias))

	mux.HandleFunc("GET /api/json", handlers.JsonExtract(medias))

	fsys, err := fs.Sub(static, "static")
	if err != nil {
		log.Fatal(err)
	}

	mux.Handle("/", http.FileServer(http.FS(fsys)))

	mux.Handle("/"+folder+"/", http.StripPrefix("/"+folder, http.FileServer(http.Dir(folder))))

	// mux.HandleFunc("/reload-templates", handlers.ReloadTemplates())

	mux.HandleFunc("GET /generate-json", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(medias)
	})

	loggedMux := middleware.LoggingMiddleware(mux)

	golog.Info("Started server http://localhost:8080")
	err = http.ListenAndServe("localhost:8080", loggedMux)
	if err != nil {
		golog.Err("%s", err.Error())
	}
}
