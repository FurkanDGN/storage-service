package config

import (
	"errors"
	"github.com/google/uuid"
	"os"
	"sync"
	"time"
	"log"
	"strconv"
)

const serverIdFilename = "server_id"

var serverId string

func init() {
	serverId = getServerId()
}

type AppConfig struct {
	MongoURL                    string
	MongoDbName                 string
	MongoVideosCollection       string
	MongoVideoServersCollection string
	VideosDir                   string
	ServerURL                   string
	ServerId                    string
	CacheRetentionTime          time.Duration
}

var Config *AppConfig

var once sync.Once

func LoadConfig() (*AppConfig, error) {
	var err error = nil

	once.Do(func() {
		mongoUrl, _ := os.LookupEnv("VIDEOHUB_MONGO_URL")
		mongoDatabaseName, _ := os.LookupEnv("VIDEOHUB_MONGO_DB_NAME")
		mongoVideosCollection, _ := os.LookupEnv("VIDEOHUB_MONGO_VIDEOS_COLLECTION")
		mongoVideoServersCollection, _ := os.LookupEnv("VIDEOHUB_MONGO_VIDEO_SERVERS_COLLECTION")
		videosDir, _ := os.LookupEnv("VIDEOHUB_VIDEOS_DIRNAME")
		serverUrl, _ := os.LookupEnv("VIDEOHUB_SERVER_URL")
		cacheRetentionTimeStr, _ := os.LookupEnv("VIDEOHUB_CACHE_RETENTION_TIME")
		cacheRetentionTimeInt, err := strconv.Atoi(cacheRetentionTimeStr)
		cacheRetentionDuration := time.Hour * time.Duration(cacheRetentionTimeInt)

		if err != nil {
    	log.Fatal(err)
		}

		if mongoUrl == "" || mongoDatabaseName == "" || mongoVideosCollection == "" || mongoVideoServersCollection == "" ||
			videosDir == "" || serverUrl == "" || cacheRetentionDuration == 0 {
			err = errors.New("missing environment variables")
			return
		}

		Config = &AppConfig{
			MongoURL:                    mongoUrl,
			MongoDbName:                 mongoDatabaseName,
			MongoVideosCollection:       mongoVideosCollection,
			MongoVideoServersCollection: mongoVideoServersCollection,
			VideosDir:                   videosDir,
			ServerURL:                   serverUrl,
			ServerId:                    serverId,
			CacheRetentionTime:          cacheRetentionDuration,
		}
	})

	return Config, err
}

func getServerId() string {
	data, err := os.ReadFile(serverIdFilename)
	var readData = ""
	if err != nil {
		readData = uuid.New().String()
		os.WriteFile(serverIdFilename, []byte(readData), 0644)
	} else {
		parsedUUID, err := uuid.Parse(string(data))
		if err != nil {
			readData = uuid.New().String()
			os.WriteFile(serverIdFilename, []byte(readData), 0644)
		} else {
			readData = parsedUUID.String()
		}
	}

	return readData
}
