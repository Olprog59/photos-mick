package handlers

import (
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"time"

	"github.com/Olprog59/golog"
	"github.com/Olprog59/photos-mick/commons"
)

func loadTemplates() (*template.Template, error) {
	funcMap := template.FuncMap{
		"dateFR": func(datetime string) string {
			t, err := time.Parse(commons.DateFormatWithSeconds, datetime)
			if err != nil {
				golog.Err("Erreur de format de date: %v", err)
				return "Date invalide"
			}
			return t.Format("Monday 02 January 2006 15:04:05")
		},
	}

	// Charger les templates à partir des fichiers intégrés
	templatesFS, err := fs.Sub(templateFS, "templates")
	if err != nil {
		return nil, err
	}

	tmpl := template.New("").Funcs(funcMap)
	err = fs.WalkDir(templatesFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			content, err := fs.ReadFile(templatesFS, path)
			if err != nil {
				return err
			}
			tmpl, err = tmpl.New(filepath.Base(path)).Parse(string(content))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

func handleGeneric(w http.ResponseWriter, data any, mainTemplate string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl, err := loadTemplates()
	if err != nil {
		golog.Err("Erreur lors du chargement des templates: %v", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	// Vérifier si le template principal existe
	if tmpl.Lookup(mainTemplate) == nil {
		golog.Err("Le template principal '%s' n'existe pas", mainTemplate)
		http.Error(w, "Template non trouvé", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, mainTemplate, data)
	if err != nil {
		golog.Err("Erreur lors de l'exécution du template %s: %v", mainTemplate, err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}
