package exiftool

import "time"

type Metadata struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Date        time.Time `json:"date"`
	Zone        string    `json:"zone"`
	Description string    `json:"description"`
	Lon         float64   `json:"lon"`
	Lat         float64   `json:"lat"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	Duration    int       `json:"duration"`
}
