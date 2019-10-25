package worker

import (
	"context"
	"crontab/crontab/common"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
	"log"
	"time"
)

//任务管理器
type JobMgr struct {
	client  *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease
	watcher clientv3.Watcher
}

var (
	//单例
	G_jobMgr *JobMgr
)

//监听任务变化
func (jobMgr *JobMgr) watchJobs() (err error) {
	var (
		getResp            *clientv3.GetResponse
		kvPair             *mvccpb.KeyValue
		job                *common.Job
		jobName            string
		jobEvent           *common.JobEvent
		watchStartRevision int64
		watchChan          clientv3.WatchChan
		watchResp          clientv3.WatchResponse
		watchEvent         *clientv3.Event
	)

	//1.get一下/cron/jobs目录下的所有任务，获知当前集群的revision
	if getResp, err = jobMgr.kv.Get(context.TODO(), common.JobSaveDir, clientv3.WithPrefix()); err != nil {
		return
	}
	//查看当前任务
	for _, kvPair = range getResp.Kvs {
		//反序列化json得到Job
		if job, err = common.UnpackJob(kvPair.Value); err == nil {
			jobEvent = common.BuildJobEvent(common.JoBEventSave, job)
			//把这个job同步给scheduler
			G_scheduler.PushJobEvent(jobEvent)
		}
	}

	//2.从该revision向后监听变化事件
	go func() { //监听协程
		//从get时刻的后续版本开始监听变化
		watchStartRevision = getResp.Header.Revision + 1
		//监听/cron/jobs/目录的后续变化
		watchChan = jobMgr.watcher.Watch(context.TODO(), common.JobSaveDir, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
		//接收watchResp
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: //任务保存事件
					//反序列化，
					if job, err = common.UnpackJob(watchEvent.Kv.Value); err != nil {
						continue
					}
					//构建一个更新Event
					jobEvent = common.BuildJobEvent(common.JoBEventSave, job)
				case mvccpb.DELETE: //任务删除事件
					jobName = common.ExtractJobName(string(watchEvent.Kv.Key))
					//构建一个删除Event
					jobEvent = common.BuildJobEvent(common.JobEventDelete, job)
				}
				//推送给scheduler
				G_scheduler.PushJobEvent(jobEvent)
			}
		}
	}()
	return
}

func (jobMgr *JobMgr) watchKill() (err error) {
	var (
		watchChan  clientv3.WatchChan
		watchResp  clientv3.WatchResponse
		watchEvent *clientv3.Event
		jobEvent   *common.JobEvent
		jobName    string
		job        *common.Job
	)
	go func() { //监听协程
		//监听/cron/kill/目录的后续变化
		watchChan = jobMgr.watcher.Watch(context.TODO(), common.JobKillerDir, clientv3.WithPrefix())
		//接收watchResp
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: //杀死某个任务事件
					jobName = common.ExtractKillName(string(watchEvent.Kv.Key))
					job = &common.Job{
						Name: jobName,
					}
					jobEvent = common.BuildJobEvent(common.JobEventKill, job)
					G_scheduler.PushJobEvent(jobEvent)
				case mvccpb.DELETE: //任务删除事件

				}
				//推送给scheduler
				G_scheduler.PushJobEvent(jobEvent)
			}
		}
	}()
	return
}

//初始化管理器
func InitJobMgr() (err error) {
	var (
		client  *clientv3.Client
		config  clientv3.Config
		lease   clientv3.Lease
		kv      clientv3.KV
		watcher clientv3.Watcher
	)
	//初始化配置
	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndPotints,
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond,
	}
	//建立连接
	if client, err = clientv3.New(config); err != nil {
		return
	}
	//通过client得到KV和lease的API子集(client.new)
	lease = clientv3.NewLease(client)
	kv = clientv3.NewKV(client)
	watcher = clientv3.NewWatcher(client)
	//赋值单例
	G_jobMgr = &JobMgr{
		client:  client,
		kv:      kv,
		lease:   lease,
		watcher: watcher,
	}
	//启动任务监听
	if err = G_jobMgr.watchJobs(); err != nil {
		log.Fatal(err)
	}
	//启动强杀任务监听
	if err = G_jobMgr.watchKill(); err != nil {
		log.Fatal(err)
	}
	return
}

//创建任务集群锁
func (jobMgr *JobMgr) CreteJobLock(jobName string) (jobLock *JobLock) {
	jobLock = InitJobLock(jobName, jobMgr.kv, jobMgr.lease)
	return
}
