package main

import (
	"videohub/handler"
	"videohub/util"
	"videohub/config"
	"log"
	"net/http"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	initializeConfig()
	mongoDb := connectToDB()
	createVideoDirectory()
	uploadHandler, videoHandler := initializeHandlers(mongoDb)
	startServer(uploadHandler, videoHandler)
}

func initializeConfig() {
	config.LoadConfig()
	log.Println("Config loaded")
}

func connectToDB() *util.MongoDB {
	log.Println("Connecting to MongoDB")
	mongoDB, err := util.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB")
	return mongoDB;
}

func createVideoDirectory() {
	videosDir := config.Config.VideosDir
	util.CreateDir(videosDir)
}

func initializeHandlers(mongoDb *util.MongoDB) (*handler.UploadHandler, *handler.VideoHandler) {
	uploadHandler := &handler.UploadHandler{MongoDb: mongoDb}
	videoHandler := &handler.VideoHandler{MongoDb: mongoDb}
	return uploadHandler, videoHandler
}

func startServer(uploadHandler *handler.UploadHandler, videoHandler *handler.VideoHandler) {
	http.HandleFunc("/upload", wrapHandlerWithCORS(uploadHandler, "PUT"))
	http.HandleFunc("/videos", wrapHandlerWithCORS(videoHandler, "GET"))
	http.HandleFunc("/"+config.Config.VideosDir+"/", serveVideoFiles)
	http.HandleFunc("/", http.NotFound)

	log.Println("Server starting at :8080")
	http.ListenAndServe(":8080", nil)
}

func wrapHandlerWithCORS(handler http.Handler, allowedMethods string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w, allowedMethods)
		handler.ServeHTTP(w, r)
	}
}

func serveVideoFiles(w http.ResponseWriter, r *http.Request) {
	videosDir := config.Config.VideosDir
	http.StripPrefix("/"+videosDir+"/", http.FileServer(http.Dir("./"+videosDir+"/"))).ServeHTTP(w, r)
}

func enableCors(w *http.ResponseWriter, allowedMethods string) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", allowedMethods)
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
}