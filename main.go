package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"videohub/config"
	"videohub/handler"
	"videohub/util"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 0, "Port number")

	flag.Parse()

	if port == 0 {
		log.Printf("Error: 'port' argument is required.\n")
		os.Exit(1)
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	initializeConfig()
	mongoDb := connectToDB()
	uploadHandler, videosHandler, videoHandler := initializeHandlers(mongoDb)
	startServer(&port, uploadHandler, videosHandler, videoHandler)
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
	return mongoDB
}

func initializeHandlers(mongoDb *util.MongoDB) (*handler.UploadHandler, *handler.VideosHandler, *handler.VideoHandler) {
	uploadHandler := &handler.UploadHandler{MongoDb: mongoDb}
	videosHandler := &handler.VideosHandler{MongoDb: mongoDb}
	videoHandler := &handler.VideoHandler{MongoDb: mongoDb}
	return uploadHandler, videosHandler, videoHandler
}

func startServer(port *int, uploadHandler *handler.UploadHandler, videosHandler *handler.VideosHandler, videoHandler *handler.VideoHandler) {
	http.HandleFunc("/upload", wrapHandlerWithCORS(uploadHandler, "PUT"))
	http.HandleFunc("/videos", wrapHandlerWithCORS(videosHandler, "GET"))
	http.HandleFunc("/video/", wrapHandlerWithCORS(videoHandler, "GET"))
	http.HandleFunc("/", http.NotFound)

	portStr := strconv.Itoa(*port)
	log.Println("Server starting at :" + portStr)
	err := http.ListenAndServe(":"+portStr, nil)
	if err != nil {
		log.Fatalf("An error occurred when starting server: %s\n", err)
	}
}

func wrapHandlerWithCORS(handler http.Handler, allowedMethods string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w, allowedMethods)
		handler.ServeHTTP(w, r)
	}
}

func enableCors(w *http.ResponseWriter, allowedMethods string) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", allowedMethods)
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
}
