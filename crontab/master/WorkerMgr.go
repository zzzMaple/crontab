package master

import (
	"context"
	"crontab/crontab/common"
	"go.etcd.io/etcd/clientv3"
	"time"
)

type WorkerMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
}

var (
	G_workerMgr *WorkerMgr
)

func (workerMgr *WorkerMgr) workerList() (addrs []string, err error) {
	var (
		getResp  *clientv3.GetResponse
		workerIP string
	)
	addrs = make([]string, 0)
	//获取所有kv
	if getResp, err = workerMgr.kv.Get(context.TODO(), common.JobWorkerDir, clientv3.WithPrefix()); err != nil {
		return
	}
	//获取每个IP
	for _, kv := range getResp.Kvs {
		workerIP = common.ExtractWorkerIP(string(kv.Key))
		addrs = append(addrs, workerIP)
	}
	return
}

func InitWorkerMgr() (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
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
	kv = clientv3.NewKV(client)
	G_workerMgr = &WorkerMgr{
		client: client,
		kv:     kv,
	}
	return
}
