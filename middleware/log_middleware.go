package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware est un middleware qui enregistre les informations sur chaque requête HTTP
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Créez un ResponseWriter personnalisé pour capturer le code de statut
		wrappedWriter := &responseWriter{ResponseWriter: w, status: 200}

		// Appelez le gestionnaire suivant
		next.ServeHTTP(wrappedWriter, r)

		// Calculez la durée de la requête
		duration := time.Since(start)

		// Enregistrez les informations de la requête
		log.Printf(
			"[%s] %s %s %d %v",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			wrappedWriter.status,
			duration,
		)
	})
}

// responseWriter est un wrapper pour http.ResponseWriter qui capture le code de statut
type responseWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader capture le code de statut avant de l'écrire
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
