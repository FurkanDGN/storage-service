package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"videohub/config"
	"videohub/util"
)

const BufferSize = 1024 * 1024 * 2

type VideoHandler struct {
	MongoDb *util.MongoDB
}

func (v *VideoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	videoID := getVideoID(r)
	video, err := v.MongoDb.FetchVideoByID(videoID)
	if err != nil {
		http.Error(w, "Video not found", http.StatusNotFound)
		return
	}

	targetPath := fmt.Sprintf("%s/%s", config.Config.VideosDir, video.VideoUrl)
	file, fileSize, err := openAndStatFile(targetPath)
	if err != nil {
		log.Printf("An error occurred: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()

	contentType := getContentType(video.VideoUrl)

	if len(r.Header.Get("Range")) == 0 {
		serveFullFile(w, r, file, contentType, fileSize)
	} else {
		servePartialFile(w, r, file, contentType, fileSize)
	}
}

func getVideoID(r *http.Request) string {
	videoID := strings.TrimPrefix(r.URL.Path, "/video/")
	return strings.TrimSuffix(videoID, filepath.Ext(videoID))
}

func openAndStatFile(targetPath string) (*os.File, int, error) {
	file, err := os.Open(targetPath)
	if err != nil {
		return nil, 0, err
	}

	fi, err := file.Stat()
	if err != nil {
		return nil, 0, err
	}

	return file, int(fi.Size()), nil
}

func getContentType(videoUrl string) string {
	fileExtension := strings.Split(videoUrl, "/")[0]
	return util.GetVideoContentType(fileExtension)
}

func serveFullFile(w http.ResponseWriter, r *http.Request, file *os.File, contentType string, fileSize int) {
	setFullFileHeaders(w, contentType, fileSize)
	sendContent(w, r, file)
}

func setFullFileHeaders(w http.ResponseWriter, contentType string, fileSize int) {
	contentLength := strconv.Itoa(fileSize)
	contentEnd := strconv.Itoa(fileSize - 1)

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Length", contentLength)
	w.Header().Set("Content-Range", fmt.Sprintf("bytes 0-%s/%s", contentEnd, contentLength))
	w.WriteHeader(200)
}

func sendContent(w http.ResponseWriter, r *http.Request, file *os.File) {
	buffer := make([]byte, BufferSize)

	for {
		select {
		case <-r.Context().Done():
			return
		default:
			n, err := file.Read(buffer)
			if err != nil || n == 0 {
				return
			}

			data := buffer[:n]
			w.Write(data)
			w.(http.Flusher).Flush()
		}
	}
}

func servePartialFile(w http.ResponseWriter, r *http.Request, file *os.File, contentType string, fileSize int) {
	rangeParam := r.Header.Get("Range")
	splitParams := strings.Split(strings.Split(rangeParam, "=")[1], "-")

	contentStartValue, _ := strconv.Atoi(splitParams[0])
	contentEndValue := fileSize - 1
	if len(splitParams) > 1 {
		val, err := strconv.Atoi(splitParams[1])
		if err == nil {
			contentEndValue = val
		}
	}

	contentLength := contentEndValue - contentStartValue + 1

	setHeaders(w, contentType, contentLength, contentStartValue, contentEndValue, fileSize)

	file.Seek(int64(contentStartValue), 0)
	sendPartialContent(w, r, file, contentEndValue)
}

func setHeaders(w http.ResponseWriter, contentType string, contentLength, contentStart, contentEnd, fileSize int) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Length", strconv.Itoa(contentLength))
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", contentStart, contentEnd, fileSize))
	w.WriteHeader(206)
}

func sendPartialContent(w http.ResponseWriter, r *http.Request, file *os.File, contentEndValue int) {
	buffer := make([]byte, BufferSize)
	writeBytes := 0

	for {
		select {
		case <-r.Context().Done():
			return
		default:
			n, err := file.Read(buffer)
			if err != nil || n == 0 {
				return
			}

			writeBytes += n
			data := buffer[:n]

			if writeBytes >= contentEndValue {
				data = buffer[:BufferSize-writeBytes+contentEndValue+1]
			}

			w.Write(data)
			w.(http.Flusher).Flush()

			if writeBytes >= contentEndValue {
				return
			}
		}
	}
}
