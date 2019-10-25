package worker

import (
	"context"
	"crontab/crontab/common"
	"go.etcd.io/etcd/clientv3"
	"net"
	"time"
)

type Register struct {
	client  *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease //租约，worker宕机时使租约过期
	localIP string         //local ip address
}

var (
	G_register *Register
)

//获取本机网卡IP
func getLocalIP() (ipv4 string, err error) {
	var (
		addrs   []net.Addr
		addr    net.Addr
		ipNet   *net.IPNet //ip地址
		isIPNet bool
	)
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return
	}
	//获取第一个非loopback的网卡IP
	for _, addr = range addrs {
		if ipNet, isIPNet = addr.(*net.IPNet); isIPNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String()
				return
			}
		}
	}
	err = common.ERR_NO_LOCAL_IP_FOUND
	return
}

//注册到/cron/workers/IP,并自动续租
func (register *Register) keepOnline() {
	var (
		regKey             string
		leaseGrantResponse *clientv3.LeaseGrantResponse
		err                error
		keepAliveChan      <-chan *clientv3.LeaseKeepAliveResponse
		keepAliveResponse  *clientv3.LeaseKeepAliveResponse
		ctx                context.Context
		cancelFunc         context.CancelFunc
	)
	for {
		//注册路径
		regKey = common.JobWorkerDir + register.localIP
		//创建租约
		if leaseGrantResponse, err = register.lease.Grant(context.TODO(), 10); err != nil {
			goto RETRY
		}
		//自动续租
		if keepAliveChan, err = register.lease.KeepAlive(context.TODO(), leaseGrantResponse.ID); err != nil {
			goto RETRY
		}
		ctx, cancelFunc = context.WithCancel(context.TODO())
		//注册到etcd中
		if _, err = register.kv.Put(ctx, regKey, "", clientv3.WithLease(leaseGrantResponse.ID)); err != nil {
			goto RETRY
		}

		for {
			select {
			case keepAliveResponse = <-keepAliveChan:
				if keepAliveResponse == nil {
					goto RETRY
				}
			}
		}
	RETRY:
		time.Sleep(1 * time.Second)
		if cancelFunc != nil {
			cancelFunc()
		}
	}
}

//初始化
func InitRegister() (err error) {
	var (
		client  *clientv3.Client
		config  clientv3.Config
		lease   clientv3.Lease
		kv      clientv3.KV
		localIP string
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
	//本机IP
	if localIP, err = getLocalIP(); err != nil {
		return
	}
	//通过client得到KV和lease的API子集(client.new)
	lease = clientv3.NewLease(client)
	kv = clientv3.NewKV(client)
	G_register = &Register{
		client:  client,
		kv:      kv,
		lease:   lease,
		localIP: localIP,
	}
	//服务注册
	go G_register.keepOnline()
	return
}
