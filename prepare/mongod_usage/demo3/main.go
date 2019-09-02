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
	JobName   string    `bson:"jobName"`
	Command   string    `bson:"command"`
	Err       string    `bson:"err"`
	Content   string    `bson:"content"`
	TimePoint TimePoint `bson:"timePoint"`
}

// jobName filter
type FindByJobName struct {
	JobName string `bson:"jobName"` //make jobName value equal to 10
}

func create(x int64) *int64 {
	return &x
}

func main() {
	var (
		client     *mongo.Client
		ctx        context.Context
		cancelFunc context.CancelFunc
		collection *mongo.Collection
		err        error
		record     *LogRecord
		cond       *FindByJobName
		cursor     *mongo.Cursor
		skipNum    *int64
		limitNum    *int64
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
	//use collection
	record = &LogRecord{
		JobName:   "job10",
		Command:   "echo hello",
		Err:       "",
		Content:   "hello",
		TimePoint: TimePoint{StartTime: time.Now().Unix(), EndTime: time.Now().Unix() + 10},
	}
	collection = client.Database("cron").Collection("job")

	cond = &FindByJobName{"job10"}

	//query
	skipNum, limitNum  = new(int64), new(int64)   //beginning num and limitation num of find query
	*skipNum, *limitNum = 0, 2
	if cursor, err = collection.Find(context.Background(), cond, &options.FindOptions{Skip: skipNum}, &options.FindOptions{Limit: limitNum}); err != nil {
		fmt.Println(err)
		return
	}

	//traverse the cursor
	for cursor.Next(context.Background()) {
		record = &LogRecord{}
		//unmarshalling (bson->object)
		if err = cursor.Decode(record); err != nil {
			fmt.Println(err)
			return
		}
		//print log
		fmt.Print(*record)
	}

	defer cursor.Close(context.Background())

}
