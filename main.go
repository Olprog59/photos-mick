package main

import (
	"log"
	"net/http"

	"github.com/Olprog59/photos-mick/handles"
	"github.com/Olprog59/photos-mick/middleware"
	"github.com/Olprog59/photos-mick/models"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	_ = handles.ReloadTemplates()
}

func main() {
	log.Println("Hello")

	medias := models.GetMetadataPhotos()

	models.ReadFileToSlice(&medias)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/medias", handles.FindAllByPage(medias))

	mux.HandleFunc("GET /media/{id}", handles.FindByID(&medias))

	mux.HandleFunc("PUT /media/{id}", handles.UpdateMedia(medias))

	mux.Handle("/", http.FileServer(http.Dir("./static")))

	mux.Handle("/medias/", http.StripPrefix("/medias", http.FileServer(http.Dir("./medias"))))

	mux.HandleFunc("/reload-templates", handles.ReloadTemplates())

	loggedMux := middleware.LoggingMiddleware(mux)

	log.Println("Started server http://localhost:8080")
	err := http.ListenAndServe("localhost:8080", loggedMux)
	if err != nil {
		log.Fatal(err)
	}
}
