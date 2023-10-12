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
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

type UploadHandler struct {
	MongoCollection *mongo.Collection
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

	// Get the file extension
	fileExtension := filepath.Ext(header.Filename)

	// Create target path
	targetPath := fmt.Sprintf("%s/%s/%s%s", config.Config.VideosDir, fileExtension, id, fileExtension)
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

	err = util.InsertVideoToDB(u.MongoCollection, video)
	if err != nil {
		log.Printf("Failed to insert video to MongoDB: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Upload successful. Video path: %s", targetPath)))
}