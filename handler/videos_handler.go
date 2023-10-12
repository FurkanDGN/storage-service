package handler

import (
	"videohub/util"
	"encoding/json"
	"net/http"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
)

type VideoHandler struct {
	MongoCollection *mongo.Collection
}

func (vh *VideoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))

	if page <= 0 || pageSize <= 0 {
		http.Error(w, "Invalid page or pageSize parameter", http.StatusBadRequest)
		return
	}

	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
	    scheme = "https"
	}

	videos, err := util.GetAllVideosFromDB(vh.MongoCollection, page, pageSize, scheme + "://" + r.Host)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	jsonResponse, err := json.Marshal(videos)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}