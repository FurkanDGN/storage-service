package util

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"path"
	"time"
	"videohub/config"
	"videohub/model"
)

type MongoDB struct {
	client                 *mongo.Client
	videosCollection       *mongo.Collection
	videoServersCollection *mongo.Collection
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
	videoServersCollection := client.Database(config.Config.MongoDbName).Collection(config.Config.MongoVideoServersCollection)
	return &MongoDB{client: client, videosCollection: videosCollection, videoServersCollection: videoServersCollection}, nil
}

func (db *MongoDB) InsertVideo(video model.Video) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": video.ID}
	update := bson.D{{"$set", video}}

	_, err := db.videosCollection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	return err
}

func (db *MongoDB) addStringToReplicatesArray(videoId string, serverId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": videoId}
	update := bson.M{"$push": bson.M{"replicates": serverId}}

	_, err := db.videosCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (db *MongoDB) GetAllVideosPaged(page int, pageSize int) ([]model.Video, error) {
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
		video.VideoUrl = config.Config.ServerURL + "/video/" + path.Base(video.VideoUrl)
		videos = append(videos, video)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return videos, nil
}

func (db *MongoDB) GetAllVideos() ([]model.Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find()

	cursor, err := db.videosCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var videos []model.Video
	for cursor.Next(ctx) {
		var video model.Video
		cursor.Decode(&video)
		video.VideoUrl = config.Config.ServerURL + "/video/" + path.Base(video.VideoUrl)
		videos = append(videos, video)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return videos, nil
}

func (db *MongoDB) FetchAllVideoServers() ([]model.VideoServer, error) {
	var servers []model.VideoServer
	cursor, err := db.videoServersCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var server model.VideoServer
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
