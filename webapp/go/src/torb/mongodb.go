package main

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	MONGO_DB_NAME         = "default"
	EVENT_COLLECTION_NAME = "events"
)

type mongoDBClient struct {
	Client *mongo.Client
}

func NewMongoDB() error {
	host := fmt.Sprintf("mongodb://%s:27017", os.Getenv("MONGODB_HOST"))
	opt := &options.ClientOptions{}
	cli, err := mongo.NewClient(opt.ApplyURI(host))
	if err != nil {
		fmt.Println("MONGODB ERROR: ", err)
		return err
	}
	cli.Connect(context.Background())
	err = cli.Ping(context.Background(), &readpref.ReadPref{})
	if err != nil {
		panic(err)
	}
	cli.Disconnect(context.Background())
	return nil
}

func FetchMongoDBClient() (*mongoDBClient, error) {
	host := fmt.Sprintf("mongodb://%s:27017", os.Getenv("MONGODB_HOST"))
	opt := &options.ClientOptions{}
	cli, err := mongo.NewClient(opt.ApplyURI(host))
	if err != nil {
		fmt.Println("MONGODB: ", err)
		return nil, err
	}

	err = cli.Connect(context.Background())
	if err != nil {
		fmt.Println("MONGODB: ", err)
		return nil, err
	}

	return &mongoDBClient{Client: cli}, nil
}

func (m *mongoDBClient) InsertEvent(event *Event) error {
	col := m.Client.Database(MONGO_DB_NAME).Collection(EVENT_COLLECTION_NAME)
	_, err := col.InsertOne(context.Background(), event)
	return err
}

// update by id
func (m *mongoDBClient) UpdateEventFg(eventID int64, publicFg, closedFg bool) error {
	col := m.Client.Database(MONGO_DB_NAME).Collection(EVENT_COLLECTION_NAME)
	filter := bson.D{{"id", eventID}}
	item := bson.D{{"$set", bson.M{"publicfg": publicFg, "closedfg": closedFg}}}
	return col.FindOneAndUpdate(context.Background(), filter, item).Decode(&bson.M{})
}

func (m *mongoDBClient) Close() error {
	return m.Client.Disconnect(context.Background())
}
