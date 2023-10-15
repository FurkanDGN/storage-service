package config

import (
	"os"
	"sync"
)

type AppConfig struct {
	MongoURL string
	MongoDbName string
	MongoVideosCollection string
	MongoFtpServersCollection string
	VideosDir string
}

var Config *AppConfig
var once sync.Once

func LoadConfig() *AppConfig {
	once.Do(func() {
		mongoUrl, _ :=              os.LookupEnv("VIDEOHUB_MONGO_URL")
		mongoDatabaseName, _ := 	  os.LookupEnv("VIDEOHUB_MONGO_DB_NAME")
		mongoVideosCollection, _ := os.LookupEnv("VIDEOHUB_MONGO_VIDEOS_COLLECTION")
		mongoFtpServersCollection, _ := os.LookupEnv("VIDEOHUB_MONGO_FTP_SERVERS_COLLECTION")
		videosDir, _ :=             os.LookupEnv("VIDEOHUB_VIDEOS_DIRNAME")

		Config = &AppConfig{
			MongoURL:              mongoUrl,
			MongoDbName:           mongoDatabaseName,
			MongoVideosCollection: mongoVideosCollection,
			MongoFtpServersCollection: mongoFtpServersCollection,
			VideosDir:             videosDir,
		}
	})

	return Config
}
