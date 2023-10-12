package main

import (
   "videohub/handler"
   "videohub/util"
   "videohub/config"
   "log"
   "net/http"
)

func main() {
   config.LoadConfig()
   log.Println("Config loaded")
   log.Println("Connecting to MongoDB")
   mongoCollection, err := util.ConnectToMongoDB()
      if err != nil {
      log.Fatalf("Failed to connect to MongoDB: %v", err)
   }
   log.Println("Connected to MongoDB")

   videosDir := config.Config.VideosDir;
   util.CreateDir(videosDir)

   uploadHandler := &handler.UploadHandler{
      MongoCollection: mongoCollection,
   }
   videoHandler := &handler.VideoHandler{
      MongoCollection: mongoCollection,
   }

   http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
      enableCors(&w, "PUT")
      uploadHandler.ServeHTTP(w, r)
   })
   http.HandleFunc("/videos", func(w http.ResponseWriter, r *http.Request) {
      enableCors(&w, "GET")
      videoHandler.ServeHTTP(w, r)
   })
   http.HandleFunc("/" + videosDir + "/", func(w http.ResponseWriter, r *http.Request) {
      enableCors(&w, "GET")
      http.StripPrefix("/" + videosDir + "/", http.FileServer(http.Dir("./" + videosDir + "/"))).ServeHTTP(w, r)
   })
   http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
      http.NotFound(w, r)
   })
   
   log.Println("Server starting at :8080")
   http.ListenAndServe(":8080", nil)
}

func enableCors(w *http.ResponseWriter, allowedMethods string) {
   (*w).Header().Set("Access-Control-Allow-Origin", "*")
   (*w).Header().Set("Access-Control-Allow-Methods", allowedMethods)
   (*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}