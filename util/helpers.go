package util

import (
	"log"
	"os"
	"strings"
)

func CreateDir(dirName string) {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err := os.MkdirAll(dirName, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func GetVideoContentType(extension string) string {
	switch strings.ToLower(extension) {
	case "mp4":
		return "video/mp4"
	case "webm":
		return "video/webm"
	case "mpeg":
		return "video/mpeg"
	case "ts":
		return "video/mp2t"
	case "ogv":
		return "video/ogg"
	case "avi":
		return "video/x-msvideo"
	default:
		return "application/octet-stream"
	}
}
