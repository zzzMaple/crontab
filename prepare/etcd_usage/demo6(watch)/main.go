package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	mvccpb2 "github.com/coreos/etcd/mvcc/mvccpb"
	"time"
)

func main() {
	var (
		config             clientv3.Config
		client             *clientv3.Client
		err                error
		kv                 clientv3.KV
		event              *clientv3.Event
		getResp            *clientv3.GetResponse
		watchStartRevision int64
		watcher            clientv3.Watcher
		watchChan          clientv3.WatchChan
		watchResp          clientv3.WatchResponse
	)
	config = clientv3.Config{
		Endpoints:   []string{"116.62.45.108:2379"},
		DialTimeout: 5 * time.Second,
	}

	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}
	//use to storage kv
	kv = clientv3.NewKV(client)
	//put and delete for watching
	go func() {
		for {
			kv.Put(context.TODO(), "/cron/jobs/job1", "I am job1")
			kv.Delete(context.TODO(), "/cron/jobs/job1")
			time.Sleep(1 * time.Second)
		}
	}()

	if getResp, err = kv.Get(context.TODO(), "/cron/job/job1"); err != nil {
		fmt.Println(err)
		return
	}

	if len(getResp.Kvs) != 0 {
		fmt.Println("value :", string(getResp.Kvs[0].Value))
	}
	watchStartRevision = getResp.Header.Revision + 1
	//new watcher
	watcher = clientv3.NewWatcher(client)
	fmt.Println("start watch from this revision:", watchStartRevision)
	//storage watchResponse
	watchChan = watcher.Watch(context.TODO(), "/cron/jobs/job1", clientv3.WithRev(watchStartRevision))
	//watch events(put and delete)
	for watchResp = range watchChan {
		for _, event = range watchResp.Events{
			switch event.Type {
			case mvccpb2.PUT:
				fmt.Println("put operation: ", string(event.Kv.Value))
			case mvccpb2.DELETE:
				fmt.Println("delete operation: ", string(event.Kv.Value))
			}
		}
	}

}
