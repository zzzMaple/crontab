package worker

import (
	"context"
	"crontab/crontab/common"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

//mongodb存储日志
type LogSink struct {
	client         *mongo.Client
	logCollection  *mongo.Collection
	logChan        chan *common.JobLog
	autoCommitChan chan *common.LogBatch
}

var (
	//单例
	G_logSink *LogSink
)

//发送日志批次
func (logSink *LogSink) saveLogBatch(batch *common.LogBatch) {
	if _, err := logSink.logCollection.InsertMany(context.TODO(), batch.Logs); err != nil {
		fmt.Println(err.Error())
		return
	}
}

//日志存储协程
func (logSink *LogSink) writeLoop() {
	var (
		log          *common.JobLog
		logBatch     *common.LogBatch //当前的batch
		commitTimer  *time.Timer
		timeoutBatch *common.LogBatch //超时的batch
	)

	for {
		select {
		case log = <-logSink.logChan:
			//写到mongodb中
			//logSink中的Collection.insertOne方法
			//log在Protocol中已经注释了bson,会自动序列化
			//每次插入需要等待mongodb的一次请求往返，可能会因网络加大耗时
			if logBatch == nil {
				//初始化批次
				logBatch = &common.LogBatch{}
				//让这个批次超时自动提交
				commitTimer = time.AfterFunc( //Timer的afterFunc方法会另起一个go协程来处理下面这个闭包回调func
					time.Duration(G_config.jobLogCommitTimeout)*time.Millisecond,
					func(batch *common.LogBatch) func() {
						return func() {
							logSink.autoCommitChan <- batch
						}
					}(logBatch),
				)
			}
			logBatch.Logs = append(logBatch.Logs, log)
			if len(logBatch.Logs) >= G_config.LogBatchSize {
				//保存日志
				logSink.saveLogBatch(logBatch)
				//发送后清空logBatch
				logBatch = nil
				//取消定时器
				commitTimer.Stop()
			}
		case timeoutBatch = <-logSink.autoCommitChan: //过期批次(1s)
			if timeoutBatch != logBatch {
				continue
			}
			//保存日志
			logSink.saveLogBatch(timeoutBatch)
			//发送后清空logBatch
			logBatch = nil
		}
	}
}

func (logSink *LogSink) Append(jobLog *common.JobLog) {
	select {
	case logSink.logChan <- jobLog:
	default:
		//队列满了就丢弃
	}
}
func InitLogSink() (err error) {
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
	G_logSink = &LogSink{
		client:         client,
		logCollection:  client.Database("my_db").Collection("my_collection"),
		logChan:        make(chan *common.JobLog, 1000),
		autoCommitChan: make(chan *common.LogBatch, 1000), //对于chan必须要初始化
	}
	//启动一下Mongodb处理协程
	go G_logSink.writeLoop()
	return
}
