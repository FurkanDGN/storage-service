package util

import (
	"videohub/model"
	"videohub/config"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
	"path"
)


type MongoDB struct {
	client     *mongo.Client
	videosCollection *mongo.Collection
	ftpServersCollection *mongo.Collection
}

func Connect() (*MongoDB, error) {
	clientOptions := options.Client().ApplyURI(config.Config.MongoURL)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	videosCollection := client.Database(config.Config.MongoDbName).Collection(config.Config.MongoVideosCollection)
	ftpServersCollection := client.Database(config.Config.MongoDbName).Collection(config.Config.MongoFtpServersCollection)
	return &MongoDB{client: client, videosCollection: videosCollection, ftpServersCollection: ftpServersCollection}, nil
}

func (db *MongoDB) InsertVideo(video model.Video) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.videosCollection.InsertOne(ctx, video)
	return err
}

func (db *MongoDB) GetAllVideos(page int, pageSize int, serverUrl string) ([]model.Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	skip := (page - 1) * pageSize
	limit := int64(pageSize)
	opts := options.Find().SetSkip(int64(skip)).SetLimit(limit)

	cursor, err := db.videosCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var videos []model.Video
	for cursor.Next(ctx) {
		var video model.Video
		cursor.Decode(&video)
		video.VideoUrl = serverUrl + "/video/" + path.Base(video.VideoUrl)
		videos = append(videos, video)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return videos, nil
}

func (db *MongoDB) FetchAllFtpServers() ([]model.FtpServer, error) {
	var servers []model.FtpServer
	cursor, err := db.ftpServersCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var server model.FtpServer
		err := cursor.Decode(&server)
		if err != nil {
			return nil, err
		}
		servers = append(servers, server)
	}
	return servers, nil
}

func (db *MongoDB) FetchVideoByID(videoID string) (*model.Video, error) {
	var video model.Video
	err := db.videosCollection.FindOne(context.TODO(), bson.M{"_id": videoID}).Decode(&video)
	if err != nil {
		return nil, err
	}
	return &video, nil
}

func (db *MongoDB) FetchFtpServerByID(serverID string) (*model.FtpServer, error) {
    var server model.FtpServer
    err := db.ftpServersCollection.FindOne(context.TODO(), bson.M{"_id": serverID}).Decode(&server)
    if err != nil {
        return nil, err
    }
    return &server, nil
}