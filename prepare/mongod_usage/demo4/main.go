package main

import (
	"context"
	"fmt"
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
	JobName   string `bson:"jobName"`
	Command   string `bson:"command"`
	Err       string `bson:"err"`
	Content   string `bson:"content"`
	TimePoint TimePoint `bson:"timePoint"`
}

type TimeBeforeCond struct {
	Before int64 `bson:"$lt:"` //{"$lt":timestamp}
}

type DeleteCond struct {
	beforeCond TimeBeforeCond `bson:"timePoint.startTime"` //{"timePoint.StartTime":{"$lt":timestamp}}
}
func main() {
	var (
		client     *mongo.Client
		ctx        context.Context
		cancelFunc context.CancelFunc
		collection *mongo.Collection
		err        error
		delCond    *DeleteCond
		delResult  *mongo.DeleteResult

	)
	//connect with mongodb host
	if client, err = mongo.NewClient(options.Client().ApplyURI("mongodb://116.62.45.108:27017")); err != nil {
		fmt.Println(err)
		return
	}
	ctx, cancelFunc = context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelFunc()
	if err = client.Connect(ctx); err != nil {
		fmt.Println(err)
		return
	}
	collection = client.Database("cron").Collection("job")

	//delete log before sometime
	//delete({"timePoint.startTime:"{"$lt":now}})
	delCond = &DeleteCond{beforeCond:TimeBeforeCond{Before:time.Now().Unix()}}

	//delete
	if delResult, err = collection.DeleteMany(context.Background(), delCond); err != nil {
		fmt.Print(err)
		return
	}
	fmt.Print("num of deleted: ",delResult.DeletedCount)
}
