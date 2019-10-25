package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

func main() {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		err     error
		kv      clientv3.KV
		getResp *clientv3.GetResponse
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

	if getResp, err = kv.Get(context.TODO(), "/cron", clientv3.WithPrefix()); err != nil {
		fmt.Println(err)
		return
	} else if getResp.Kvs != nil {
		fmt.Println(getResp.Header.Revision)
		for _, value := range getResp.Kvs {
			fmt.Println(string(value.Key))
			fmt.Println(string(value.Value))
		}

	} else {
		fmt.Println("get nil")
	}

}
