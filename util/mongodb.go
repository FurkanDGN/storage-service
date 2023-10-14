package util

import (
	"videohub/model"
	"videohub/config"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// MongoDB wraps mongo client and collection for ease of use
type MongoDB struct {
	client     *mongo.Client
	collection *mongo.Collection
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

	collection := client.Database(config.Config.MongoDbName).Collection(config.Config.MongoVideosCollection)
	return &MongoDB{client: client, collection: collection}, nil
}

func (db *MongoDB) InsertVideo(video model.Video) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := db.collection.InsertOne(ctx, video)
	return err
}

func (db *MongoDB) GetAllVideos(page int, pageSize int, serverUrl string) ([]model.Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	skip := (page - 1) * pageSize
	limit := int64(pageSize)
	opts := options.Find().SetSkip(int64(skip)).SetLimit(limit)

	cursor, err := db.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var videos []model.Video
	for cursor.Next(ctx) {
		var video model.Video
		cursor.Decode(&video)
		video.VideoUrl = serverUrl + "/" + config.Config.VideosDir + "/" + video.VideoUrl
		videos = append(videos, video)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return videos, nil
}