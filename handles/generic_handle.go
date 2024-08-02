package handles

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/Olprog59/photos-mick/commons"
)

var (
	templates     *template.Template
	templateMutex sync.RWMutex
)

func loadTemplates() (*template.Template, error) {
	funcMap := template.FuncMap{
		"dateFR": func(datetime string) string {
			t, err := time.Parse(commons.DateFormatWithSeconds, datetime)
			if err != nil {
				log.Printf("Erreur de format de date: %v", err)
				return "Date invalide"
			}
			return t.Format("Monday 02 January 2006 15:04:05")
		},
	}

	log.Println("Chargement des templates")

	globPattern := filepath.Join(commons.TemplateDir, "*.html")
	log.Printf("Chargement des templates depuis %s", globPattern)
	tmpl, err := template.New("").Funcs(funcMap).ParseGlob(globPattern)
	if err != nil {
		return nil, err
	}

	// Log des templates chargés
	for _, t := range tmpl.Templates() {
		log.Printf("Template chargé: %s", t.Name())
	}

	return tmpl, nil
}

func handleGeneric(w http.ResponseWriter, data any, mainTemplate string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := loadTemplates()
	if err != nil {
		log.Printf("Erreur lors du chargement des templates: %v", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	// Vérifier si le template principal existe
	if tmpl.Lookup(mainTemplate) == nil {
		log.Printf("Le template principal '%s' n'existe pas", mainTemplate)
		http.Error(w, "Template non trouvé", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, mainTemplate, data)
	if err != nil {
		log.Printf("Erreur lors de l'exécution du template %s: %v", mainTemplate, err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}
