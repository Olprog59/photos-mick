package commons

import (
	"log"
	"strconv"
)

func StringToFloat64(str string) float64 {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		log.Println(err)
		return 0
	}
	return f
}
