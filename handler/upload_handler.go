package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"videohub/config"
	"videohub/model"
	"videohub/util"
)

type UploadHandler struct {
	MongoDb *util.MongoDB
}

type Response struct {
	Success   bool   `json:"success"`
	VideoPath string `json:"videoPath"`
}

func (u *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to read uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	title, err := validateForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := uuid.New().String()

	targetPath, err := saveFileToDisk(file, header, id)
	if err != nil {
		log.Printf("An error occurred when saving file to disk: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	video := prepareVideoModel(id, title, targetPath, []string{})
	if err = u.MongoDb.InsertVideo(video); err != nil {
		log.Printf("An error occurred when saving file to mongoDB: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := Response{
		Success:   true,
		VideoPath: fmt.Sprintf("%s%s", config.Config.ServerURL, "/video/"+id),
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func validateForm(r *http.Request) (title string, err error) {
	title = r.FormValue("title")
	if title == "" {
		err = errors.New("title cannot be null")
	}

	return
}

func saveFileToDisk(file multipart.File, header *multipart.FileHeader, id string) (targetPath string, err error) {
	fileExtension := strings.TrimPrefix(filepath.Ext(header.Filename), ".")
	targetPath = fmt.Sprintf("%s/%s", fileExtension, id)
	filePath := fmt.Sprintf("%s/%s", config.Config.VideosDir, targetPath)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return "", err
	}
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err = io.Copy(dst, file); err != nil {
		return "", err
	}
	return targetPath, nil
}

func prepareVideoModel(id, title, targetPath string, replicates []string) model.Video {
	return model.Video{
		ID:         id,
		Title:      title,
		VideoUrl:   targetPath,
		Replicates: replicates,
	}
}
