package handler

import (
	"videohub/util"
	"net/http"
	"log"
	"io"
	"path/filepath"
	"strings"
	"strconv"
	"bytes"
)

const BUFSIZE = 1024 * 8

type VideoHandler struct {
	MongoDb *util.MongoDB
}


func (v *VideoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	videoID := strings.TrimPrefix(r.URL.Path, "/video/")
  videoID = strings.TrimSuffix(videoID, filepath.Ext(videoID))

  video, err := v.MongoDb.FetchVideoByID(videoID)
  if err != nil {
      http.Error(w, "Video not found", http.StatusNotFound)
      return
  }

  ftpServerID := video.Replicates[0]
  ftpServer, err := v.MongoDb.FetchFtpServerByID(ftpServerID)
  if err != nil {
      http.Error(w, "Internal server error", http.StatusInternalServerError)
      return
  }

	buffer, err := util.FetchVideoFromFtpToBuffer(*ftpServer, "videos/" + video.VideoUrl)
	if err != nil {
		http.Error(w, "Failed to fetch video from FTP server", http.StatusInternalServerError)
		log.Printf("Failed to fetch video: %v", err)
		return
	}

	videoSize := buffer.Len()
	contentLength := strconv.Itoa(videoSize)
	contentEnd := strconv.Itoa(videoSize - 1)

	if len(r.Header.Get("Range")) == 0 {
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Length", contentLength)
		w.Header().Set("Content-Range", "bytes 0-" + contentEnd + "/" + contentLength)
		w.WriteHeader(200)

		videoData := buffer.Bytes()
		for start := 0; start < len(videoData); {
			end := start + BUFSIZE
			if end > len(videoData) {
				end = len(videoData)
			}
			w.Write(videoData[start:end])
			w.(http.Flusher).Flush()
			start = end
		}
	} else {
		rangeParam := strings.Split(r.Header.Get("Range"), "=")[1]
		splitParams := strings.Split(rangeParam, "-")

		contentStartValue, err := strconv.Atoi(splitParams[0])
		if err != nil {
			contentStartValue = 0
		}

		contentEndValue, err := strconv.Atoi(splitParams[1])
		if err != nil {
			contentEndValue = videoSize - 1
		}

		contentStart := strconv.Itoa(contentStartValue)
		contentEnd := strconv.Itoa(contentEndValue)
		contentSize := strconv.Itoa(videoSize)
		contentLength = strconv.Itoa(contentEndValue - contentStartValue + 1)

		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Length", contentLength)
		w.Header().Set("Content-Range", "bytes " + contentStart + "-" + contentEnd + "/" + contentSize)
		w.WriteHeader(206)

		io.CopyN(w, bytes.NewReader(buffer.Bytes()[contentStartValue:]), int64(contentEndValue - contentStartValue + 1))
	}
}