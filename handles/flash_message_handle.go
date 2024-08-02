package handles

import (
	"fmt"
	"net/http"
)

func messageHeader(w *http.ResponseWriter, message string, data ...any) {
	(*w).Header().Set("Message", fmt.Sprintf(message+"\n", data...))
	(*w).Header().Set("Content-Type", "text/html; charset=utf-8")
}
