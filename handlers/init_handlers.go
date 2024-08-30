package handlers

import "io/fs"

var templateFS fs.FS

func InitHandlers(template fs.FS) {
	templateFS = template
}
