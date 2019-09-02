package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

func main() {
	var (
		config clientv3.Config
		client *clientv3.Client
		err error
		kv clientv3.KV
		putResp *clientv3.PutResponse
	)
	config = clientv3.Config{
		Endpoints:[]string{"116.62.45.108:2379"},
		DialTimeout: 5 * time.Second,
	}

	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}
	//use to storage keyvalue
	kv = clientv3.NewKV(client)
	if putResp, err = kv.Put(context.TODO(),"1112","hello1",clientv3.WithPrevKV()); err != nil {
		fmt.Println(err)
	}else {
		fmt.Println("Revision", putResp.Header.Revision)
		if putResp.PrevKv != nil {
			fmt.Println("PrevValue: ",string(putResp.PrevKv.Value))
		}
	}
}
