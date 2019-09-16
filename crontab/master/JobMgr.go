package master

import (
	"context"
	"crontab/crontab/common"
	"encoding/json"
	"go.etcd.io/etcd/clientv3"
	"time"
)

//任务管理器
type JobMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

var (
	//单例
	G_jobMgr *JobMgr
)

//初始化管理器
func InitJobMgr() (err error) {
	var (
		client *clientv3.Client
		config clientv3.Config
		lease  clientv3.Lease
		kv     clientv3.KV
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
	//赋值单例
	G_jobMgr = &JobMgr{
		client: client,
		kv:     kv,
		lease:  lease,
	}
	return
}

//JobMgr的JobSave方法，参数为job，返回oldJob和error，保存到ETCD(/cron/jobs/jobName)->json
func (jobMgr *JobMgr) JobSave(job *common.Job) (oldJob *common.Job, err error) {
	//jobKey, jobValue
	var (
		jobKey   string
		jobValue []byte
		putResp  *clientv3.PutResponse
	)
	//etcd保存key
	jobKey = "/cron/jobs/" + job.Name
	//生成任务信息Json
	if jobValue, err = json.Marshal(job); err != nil {
		goto ERR
	}
	//保存到ETCD
	if putResp, err = jobMgr.kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV()); err != nil {
		goto ERR
	}
	//若更新，返回旧值
	if putResp.PrevKv != nil {
		//对旧值进行反序列化
		if err = json.Unmarshal(putResp.PrevKv.Value, &oldJob); err != nil {
			err = nil //忽略旧值
			return
		}
	}
ERR:
	return
}

//JobMgr的JobDelete方法
func (jobMgr *JobMgr) JobDelete(jobName string) (oldJob *common.Job, err error) {
	var (
		jobKey  string
		delResp *clientv3.DeleteResponse
	)
	//etcd中保存任务的Key
	jobKey = "/cron/jobs/" + jobName
	//从etcd中删除
	if delResp, err = jobMgr.kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		goto ERR
	}
	//返回被删除的任务信息(反序列化)
	if len(delResp.PrevKvs) != 0 {
		//解析旧值，返回它
		if err = json.Unmarshal(delResp.PrevKvs[0].Value, &oldJob); err != nil {
			err = nil //即使旧值为空也OK，故给err赋空
			return
		}
	}
ERR:
	return
}
