package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

func main() {
	var (
		config       clientv3.Config
		client       *clientv3.Client
		err          error
		kv           clientv3.KV
		lease        clientv3.Lease
		leaseId      clientv3.LeaseID
		putResp      *clientv3.PutResponse
		getResp      *clientv3.GetResponse
		keepResp     *clientv3.LeaseKeepAliveResponse
		keepRespChan <-chan *clientv3.LeaseKeepAliveResponse
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
	//apply a lease

	//auto lease, every keepAlive application will extend one second of lease
	if keepRespChan, err = lease.KeepAlive(context.TODO(), leaseId); err != nil {
		fmt.Println(err)
		return
	}
	go func() {
		for {
			select {
			case keepResp = <-keepRespChan:
				if keepRespChan == nil {
					fmt.Println("lease expired")
					goto END
				} else {
					fmt.Println("receive an application for extension the lease ", leaseId)
				}

			}
		}
	END:
	}()
	//put a kv, link it with the lease we apply
	if putResp, err = kv.Put(context.TODO(), "00011", "hello", clientv3.WithLease(leaseId)); err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println(putResp.Header.Revision)
	}

	//check if the lease expired
	for {
		if getResp, err = kv.Get(context.TODO(), "00011"); err != nil {
			fmt.Println(err)
			return
		}
		if getResp.Count == 0 {
			fmt.Println("kv expired")
			break
		}
		fmt.Println("kv still alive", getResp.Kvs)
		time.Sleep(2 * time.Second)
	}

}
