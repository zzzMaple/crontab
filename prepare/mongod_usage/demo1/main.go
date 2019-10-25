package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func main() {
	var (
		client     *mongo.Client
		ctx        context.Context
		cancelFunc context.CancelFunc
		collection *mongo.Collection
		res        *mongo.InsertOneResult
		err        error
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
	collection = client.Database("my_db").Collection("my_collection")
	if res, err = collection.InsertOne(context.Background(), bson.M{"hello": "world11"}); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("id: ", res.InsertedID)
}
