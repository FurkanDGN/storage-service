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

func ConnectToMongoDB() (*mongo.Collection, error) {
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
	return collection, nil
}

func InsertVideoToDB(collection *mongo.Collection, video model.Video) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, video)
	if err != nil {
		return err
	}
	return nil
}

func GetAllVideosFromDB(collection *mongo.Collection, page int, pageSize int) ([]model.Video, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	skip := (page - 1) * pageSize
	limit := int64(pageSize)
	opts := options.Find().SetSkip(int64(skip)).SetLimit(limit)

	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var videos []model.Video
	for cursor.Next(ctx) {
		var video model.Video
		cursor.Decode(&video)
		video.VideoUrl = config.Config.ServerURL + "/" + video.VideoUrl
		videos = append(videos, video)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return videos, nil
}