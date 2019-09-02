package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

func main() {
	//lease => make lock auto update
	//Operation
	//txn --transaction
	//workflow
	//1. lock(create lease, auto apply lease, use the lease to seize a key)
	//2. transaction
	//3. free lock(cancel application for the lease, free lease)
	var (
		config         clientv3.Config
		client         *clientv3.Client
		err            error
		kv             clientv3.KV
		lease          clientv3.Lease
		leaseId        clientv3.LeaseID
		keepResp       *clientv3.LeaseKeepAliveResponse
		keepRespChan   <-chan *clientv3.LeaseKeepAliveResponse
		leaseGrantResp *clientv3.LeaseGrantResponse
		ctx            context.Context
		cancelFunc     context.CancelFunc
		txn            clientv3.Txn
		txnResp        *clientv3.TxnResponse
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

	//apply a lease
	lease = clientv3.NewLease(client)
	if leaseGrantResp, err = lease.Grant(context.TODO(),5); err != nil {
		fmt.Println(err)
		return
	}
	//get the ID
	leaseId = leaseGrantResp.ID
	//cancle func, free lease while func finished
	ctx, cancelFunc = context.WithCancel(context.TODO())
	defer cancelFunc()
	defer lease.Revoke(context.TODO(), leaseId)
	//keep alive
	if keepRespChan, err = lease.KeepAlive(ctx, leaseId); err != nil {
		fmt.Println(err)
		return
	}
	//auto keep lease alive
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


	//create transaction
	txn = kv.Txn(context.TODO())
	//if key(kv) did not exist, create key(kv)
	txn.If(clientv3.Compare(clientv3.CreateRevision("/cron/jobs/job3"),"=",0)).
		Then(clientv3.OpPut("/cron/jobs/job3","zero",clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet("/cron/jobs/job3"))
	//commit transaction
	if txnResp,err = txn.Commit(); err != nil {
		fmt.Println(err)
		return
	}

	if !txnResp.Succeeded{
		fmt.Println("failed to get the lock", txnResp.Responses[0].GetResponseRange().Kvs[0].Value)
	}
	fmt.Println("do task")
	time.Sleep(5 * time.Second)
	//free the key, => defer
}
