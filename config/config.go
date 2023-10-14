package config

import (
	"os"
	"sync"
)

type AppConfig struct {
	MongoURL string
	MongoDbName string
	MongoVideosCollection string
	ServerURL string
	VideosDir string
}

var Config *AppConfig
var once sync.Once

func LoadConfig() *AppConfig {
	once.Do(func() {
		mongoUrl, _ :=              os.LookupEnv("VIDEOHUB_MONGO_URL")
		mongoDatabaseName, _ := 	  os.LookupEnv("VIDEOHUB_MONGO_DB_NAME")
		mongoVideosCollection, _ := os.LookupEnv("VIDEOHUB_MONGO_VIDEOS_COLLECTION")
		serverURL, _ :=             os.LookupEnv("VIDEOHUB_SERVER_URL")
		videosDir, _ :=             os.LookupEnv("VIDEOHUB_VIDEOS_DIRNAME")

		Config = &AppConfig{
			MongoURL:              mongoUrl,
			MongoDbName:           mongoDatabaseName,
			MongoVideosCollection: mongoVideosCollection,
			ServerURL:             serverURL,
			VideosDir:             videosDir,
		}
	})

	return Config
}
