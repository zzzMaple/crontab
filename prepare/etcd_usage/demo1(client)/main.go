package main

import (
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

func main() {
	var (
		config clientv3.Config
		client *clientv3.Client
		err error
	)
	//setting config
	config = clientv3.Config{
		Endpoints:[]string{"116.62.45.108:2379"}, //list of cluster, could more than one
		DialTimeout: 5 * time.Second,
	}
	//client link to server
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("success")
	client = client
}
