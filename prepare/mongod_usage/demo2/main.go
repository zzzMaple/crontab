package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

//task execution timing
type TimePoint struct {
	StartTime int64 `bson:"startTime"`
	EndTime   int64 `bson:"endTime"`
}

//a piece of log
type LogRecord struct {
	JobName   string    `bson:"jobName"`
	Command   string    `bson:"command"`
	Err       string    `bson:"err"`
	Content   string    `bson:"content"`
	TimePoint TimePoint `bson:"timePoint"`
}

func main() {
	var (
		client     *mongo.Client
		ctx        context.Context
		cancelFunc context.CancelFunc
		collection *mongo.Collection
		res        *mongo.InsertOneResult
		err        error
		record     *LogRecord
		docId      primitive.ObjectID
	)
	//connect with mongodb host
	if client, err = mongo.NewClient(options.Client().ApplyURI("mongodb://116.62.45.108:27017")); err != nil {
		fmt.Println(err)
		return
	}
	ctx, cancelFunc = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	if err = client.Connect(ctx); err != nil {
		fmt.Println(err)
		return
	}
	//use collection
	record = &LogRecord{
		JobName:   "job10",
		Command:   "echo hello",
		Err:       "",
		Content:   "hello",
		TimePoint: TimePoint{StartTime: time.Now().Unix(), EndTime: time.Now().Unix() + 10},
	}
	collection = client.Database("cron").Collection("job")
	if res, err = collection.InsertOne(context.Background(), record); err != nil {
		fmt.Println(err)
		return
	}
	docId = res.InsertedID.(primitive.ObjectID)
	fmt.Println("id: ", docId.Hex())

}
