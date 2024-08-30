package handlers

import "net/http"

func IndexHandle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleGeneric(w, nil, "index")
	}
}
