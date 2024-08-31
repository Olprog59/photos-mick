package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"
)

// func messageHeader(w *http.ResponseWriter, message string, data ...any) {
// 	(*w).Header().Set("Message", fmt.Sprintf(message+"\n", data...))
// 	(*w).Header().Set("Content-Type", "text/html; charset=utf-8")
// }

func messageHeader(w *http.ResponseWriter, message string, data ...any) {
	formattedMessage := fmt.Sprintf(message, data...)

	// Assurez-vous que formattedMessage est en UTF-8
	utf8Message := []byte(formattedMessage)

	// Encodez en Base64
	encodedMessage := base64.StdEncoding.EncodeToString(utf8Message)

	(*w).Header().Set("Message", encodedMessage)
	(*w).Header().Set("Content-Type", "text/html; charset=utf-8")
}
