package master

import (
	"context"
	"crontab/crontab/common"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

//mongodb日志

type LogMgr struct {
	client        *mongo.Client
	logCollection *mongo.Collection
}

var (
	G_logMgr *LogMgr
)

//查询log
func (logMgr *LogMgr) ListLog(name string, skip int, limit int) (logArr []*common.JobLog, err error) {
	var (
		filter  *common.JobLogFilter       //过滤器
		logSort *common.SortLogByStartTime //排序方式
		findops *options.FindOptions       //查询参数
		cursor  *mongo.Cursor              //游标
		jobLog  *common.JobLog
	)

	logArr = make([]*common.JobLog, 0) //初始化，如果find的return为空则直接返回该logArr

	//过滤条件
	filter = &common.JobLogFilter{
		JobName: name,
	}

	//按照任务开始时间倒排
	logSort = &common.SortLogByStartTime{SortOrder: -1}
	//设置查询参数
	findops = &options.FindOptions{Sort: logSort}
	findops.SetSkip(int64(skip))
	findops.SetLimit(int64(limit))

	if cursor, err = logMgr.logCollection.Find(context.TODO(), filter, findops); err != nil {
		return
	}
	//延迟释放游标
	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO()) {
		jobLog = &common.JobLog{}

		//反序列化BSON
		if err = cursor.Decode(jobLog); err != nil {
			continue
		}
		logArr = append(logArr, jobLog)
	}
	return
}
func InitLogMgr() (err error) {
	var (
		client     *mongo.Client
		ctx        context.Context
		cancelFunc context.CancelFunc
	)
	//connect with mongodb host
	if client, err = mongo.NewClient(options.Client().ApplyURI(G_config.MongodbUri)); err != nil {
		fmt.Println(err)
		return
	}
	ctx, cancelFunc = context.WithTimeout(context.Background(), time.Duration(G_config.MongodbTimeout)*time.Millisecond)
	defer cancelFunc()
	if err = client.Connect(ctx); err != nil {
		fmt.Println(err)
		return
	}
	G_logMgr = &LogMgr{
		client:        client,
		logCollection: client.Database("my_db").Collection("my_collection"),
	}
	return
}
