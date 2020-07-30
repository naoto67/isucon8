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
	MONGO_DB_NAME          = "default"
	EVENT_COLLECTION_NAME  = "events"
	REPORT_COLLECTION_NAME = "reports"
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

func (m *mongoDBClient) FindAllEvents() ([]*Event, error) {
	var events []*Event
	col := m.Client.Database(MONGO_DB_NAME).Collection(EVENT_COLLECTION_NAME)
	filter := bson.D{} // fetch all
	cursor, err := col.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	err = cursor.All(context.Background(), &events)
	return events, err
}

func (m *mongoDBClient) FindEventByID(eventID int64) (*Event, error) {
	col := m.Client.Database(MONGO_DB_NAME).Collection(EVENT_COLLECTION_NAME)
	filter := bson.D{{"id", eventID}}
	var event *Event
	err := col.FindOne(context.Background(), filter).Decode(&event)
	if err != nil {
		return nil, err
	}
	return event, err
}

func (m *mongoDBClient) Close() error {
	return m.Client.Disconnect(context.Background())
}

func (m *mongoDBClient) Truncate(collectionName string) error {
	col := m.Client.Database(MONGO_DB_NAME).Collection(collectionName)
	_, err := col.DeleteMany(context.Background(), bson.D{})
	return err
}

func (m *mongoDBClient) InsertReport(report Report) error {
	col := m.Client.Database(MONGO_DB_NAME).Collection(REPORT_COLLECTION_NAME)
	_, err := col.InsertOne(context.Background(), report)
	return err
}

func (m *mongoDBClient) BulkInsertReports(reports []interface{}) error {
	col := m.Client.Database(MONGO_DB_NAME).Collection(REPORT_COLLECTION_NAME)
	_, err := col.InsertMany(context.Background(), reports)
	return err
}

func (m *mongoDBClient) UpdateCanceledAtReport(reservationID int64, canceledAt string) error {
	col := m.Client.Database(MONGO_DB_NAME).Collection(REPORT_COLLECTION_NAME)
	filter := bson.D{{"reservationid", reservationID}}
	item := bson.D{{"$set", bson.M{"canceledat": canceledAt}}}
	return col.FindOneAndUpdate(context.Background(), filter, item).Decode(&bson.M{})
}

func (m *mongoDBClient) FindAllReports() ([]Report, error) {
	var reports []Report
	col := m.Client.Database(MONGO_DB_NAME).Collection(REPORT_COLLECTION_NAME)
	sortOption := options.Find().SetSort(bson.D{{"soldat", 1}})
	filter := bson.D{{}}
	cursor, err := col.Find(context.Background(), filter, sortOption)
	if err != nil {
		return nil, err
	}
	err = cursor.All(context.Background(), &reports)
	return reports, err
}

func (m *mongoDBClient) FindReportsByEventID(eventID int64) ([]Report, error) {
	var reports []Report
	col := m.Client.Database(MONGO_DB_NAME).Collection(REPORT_COLLECTION_NAME)
	sortOption := options.Find().SetSort(bson.D{{"soldat", 1}})
	filter := bson.D{{"eventid", eventID}}
	cursor, err := col.Find(context.Background(), filter, sortOption)
	if err != nil {
		return nil, err
	}
	err = cursor.All(context.Background(), &reports)
	return reports, err
}
