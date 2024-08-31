package commons

import (
	"strconv"
	"time"

	"github.com/Olprog59/golog"
)

func StringToFloat64(str string) float64 {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		golog.Err("Error converting string to float64: %v", err)
		return 0
	}
	return f
}

func ConvertISOToGPSTime(isoTime string) (string, error) {
	// Définir le format de la date ISO 8601
	isoTimeFormat := ExifDateFormatWithSecondsWithTimeZone

	// Parser le temps ISO 8601
	t, err := time.Parse(isoTimeFormat, isoTime)
	if err != nil {
		return "", err
	}

	// Convertir en format GPS Date/Time
	gpsTime := t.UTC().Format("2006:01:02 15:04:05Z")
	return gpsTime, nil
}

func ConvertISOToExifTime(isoTime string) (string, error) {
	// Définir le format de la date ISO 8601
	isoTimeFormat := DateFormatWithSeconds

	// Parser le temps ISO 8601
	t, err := time.Parse(isoTimeFormat, isoTime)
	if err != nil {
		return "", err
	}

	// Convertir en format Exif (YYYY:MM:DD HH:MM:SS)
	exifTime := t.Format("2006:01:02 15:04:05")
	return exifTime, nil
}
