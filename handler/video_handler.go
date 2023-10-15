package handler

import (
	"videohub/util"
	"net/http"
	"log"
	"io"
	"path/filepath"
	"strings"
	"strconv"
	"fmt"
)

const BUFSIZE = 1024 * 1024

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

	videoSize, err := util.FetchVideoSizeFromFtp(*ftpServer, "videos/" + video.VideoUrl)
	if err != nil {
		log.Printf("Failed to fetch video size: %v", err)
		http.Error(w, "Failed to fetch video size", http.StatusInternalServerError)
		return
	}

	contentLength := strconv.FormatInt(videoSize, 10)

	if len(r.Header.Get("Range")) == 0 {
		conn, videoReader, err := util.FetchVideoReaderAndConnFromFtp(*ftpServer, "videos/"+video.VideoUrl)
    if err != nil {
        http.Error(w, "Failed to fetch video from FTP server", http.StatusInternalServerError)
        log.Printf("Failed to fetch video: %v", err)
        return
    }
    defer conn.Quit()
    defer videoReader.Close()

    w.Header().Set("Content-Type", "video/mp4")
    w.Header().Set("Accept-Ranges", "bytes")
    w.Header().Set("Content-Length", contentLength)
    w.WriteHeader(200)

		buf := make([]byte, BUFSIZE)
    io.CopyBuffer(w, videoReader, buf)
	} else {
		rangeParam := strings.Split(r.Header.Get("Range"), "=")[1]
    splitParams := strings.Split(rangeParam, "-")
    
    var contentEndValue int64
		var contentStartValue int64
		var err error

		tempValue, err := strconv.Atoi(splitParams[0])
		if err != nil {
		    contentStartValue = 0
		} else {
		    contentStartValue = int64(tempValue)
		}

		tempValue, err = strconv.Atoi(splitParams[1])
		if err != nil {
		    contentEndValue = videoSize - 1
		} else {
		    contentEndValue = int64(tempValue)
		}

		conn, partialReader, err := util.FetchPartialVideoFromFtpToReader(*ftpServer, "videos/"+video.VideoUrl, contentStartValue, contentEndValue)
    if err != nil {
        http.Error(w, "Failed to fetch video from FTP server", http.StatusInternalServerError)
        log.Printf("Failed to fetch video: %v", err)
        return
    }
    defer conn.Quit()

    w.Header().Set("Content-Type", "video/mp4")
    w.Header().Set("Accept-Ranges", "bytes")
    w.Header().Set("Content-Length", strconv.Itoa(int(contentEndValue - contentStartValue + 1)))
    w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", contentStartValue, contentEndValue, videoSize))
    w.WriteHeader(206)

    io.Copy(w, partialReader)
	}
}