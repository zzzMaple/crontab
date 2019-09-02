package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

func main() {
	var (
		config clientv3.Config
		client *clientv3.Client
		err    error
		kv     clientv3.KV
		putOp  clientv3.Op
		getOp  clientv3.Op
		opResp clientv3.OpResponse
	)
	config = clientv3.Config{
		Endpoints:   []string{"116.62.45.108:2379"},
		DialTimeout: 5 * time.Second,
	}

	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}
	//use to storage keyvalue
	kv = clientv3.NewKV(client)
	//op:Operation,put op
	putOp = clientv3.OpPut("/cron/jobs/job2","hello, Op")

	if opResp, err = kv.Do(context.TODO(),putOp); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("put success, revision:", opResp.Put().Header.Revision)
	//get op
	getOp = clientv3.OpGet("/cron/jobs/job2")
	if opResp, err = kv.Do(context.TODO(),getOp); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("get success, revision: ",opResp.Get().Header.Revision)
	fmt.Println("key: ", string(opResp.Get().Kvs[0].Key), "Value: ", string(opResp.Get().Kvs[0].Value))




}
