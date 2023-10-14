package handler

import (
	"videohub/util"
	"videohub/model"
	"videohub/config"
	"path/filepath"
	"io"
	"fmt"
	"net/http"
	"os"
	"log"
	"strings"
)

type UploadHandler struct {
	MongoDb *util.MongoDB
}

func (u *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to read uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	id := r.FormValue("id")
	title := r.FormValue("title")

	if title == "" {
		http.Error(w, "Title cannot be null", http.StatusBadRequest)
		return
	}

	fileExtension := strings.TrimPrefix(filepath.Ext(header.Filename), ".")
	targetPath := fmt.Sprintf("%s/%s/%s%s", config.Config.VideosDir, fileExtension, id, filepath.Ext(header.Filename))
	err = os.MkdirAll(filepath.Dir(targetPath), 0755)
	if err != nil {
		http.Error(w, "Failed to create directory", http.StatusInternalServerError)
		return
	}

	targetFile, err := os.Create(targetPath)
	if err != nil {
		http.Error(w, "Failed to create target file", http.StatusInternalServerError)
		return
	}
	defer targetFile.Close()

	_, err = io.Copy(targetFile, file)
	if err != nil {
		http.Error(w, "Failed to copy to target file", http.StatusInternalServerError)
		return
	}

	video := model.Video{
		ID:    id,
		Title: title,
		VideoUrl: fmt.Sprintf("%s/%s.%s", fileExtension, id, fileExtension),
	}

	err = u.MongoDb.InsertVideo(video)
	if err != nil {
		log.Printf("Failed to insert video to MongoDB: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		 scheme = "https"
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Upload successful. Video path: %s", scheme + "://" + r.Host + "/" + targetPath)))
}