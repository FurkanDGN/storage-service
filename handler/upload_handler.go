package handler

import (
	"videohub/util"
	"videohub/model"
	"videohub/config"
	"path/filepath"
	"io"
	"fmt"
	"net/http"
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

	ftpServers, err := u.MongoDb.FetchAllFtpServers()
	if err != nil {
		log.Printf("Failed to fetch FTP servers: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var replicates []string
	for _, server := range ftpServers {
		replicates = append(replicates, server.ID)
	}

	video := model.Video{
		ID:        id,
		Title:     title,
		VideoUrl:  fmt.Sprintf("%s/%s.%s", fileExtension, id, fileExtension),
		Replicates: replicates,
	}

	err = u.MongoDb.InsertVideo(video)
	if err != nil {
		log.Printf("Failed to insert video to MongoDB: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	u.UploadToFtpServers(file, targetPath)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Upload successful. Video path: %s", "https://ultrahqporn.com/video/" + id)))
}

func (u *UploadHandler) UploadToFtpServers(file io.ReadSeeker, ftpPath string) {
	ftpServers, err := u.MongoDb.FetchAllFtpServers()
	if err != nil {
		log.Printf("Failed to fetch FTP servers: %v", err)
		return
	}

	for _, server := range ftpServers {
		conn, err := util.ConnectToFtp(server)
		if err != nil {
			log.Printf("Failed to connect to FTP server %s: %v", server.Address, err)
			continue
		}
		defer conn.Quit()

    directory := filepath.Dir(ftpPath)
    err = util.EnsureFtpDirectories(conn, directory)
    if err != nil {
        log.Printf("Failed to create FTP directory %s on server %s: %v", directory, server.Address, err)
        continue
    }

		file.Seek(0, 0)
		err = conn.Stor(ftpPath, file)
		if err != nil {
			log.Printf("Failed to upload file to FTP server %s: %v", server.Address, err)
		}
	}
}