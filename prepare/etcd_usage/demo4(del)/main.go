package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"time"
)

func main() {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		err     error
		kv      clientv3.KV
		delResp *clientv3.DeleteResponse
		kvpair  *mvccpb.KeyValue
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

	//delete kv
	if delResp, err = kv.Delete(context.TODO(), "111", clientv3.WithPrevKV()); err != nil {
		fmt.Println(err)
		return
	}
	//kvpair include Key and Value
	if len(delResp.PrevKvs) != 0 {
		for _, kvpair = range delResp.PrevKvs {
			fmt.Println("delete Key: ", string(kvpair.Key), "value: ",string(kvpair.Value))
		}
	}else{
		fmt.Println("didnt delete anything")
	}

}
