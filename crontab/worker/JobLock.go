package worker

import (
	"context"
	"crontab/crontab/common"
	"go.etcd.io/etcd/clientv3"
	"log"
)

//分布式锁(TXN事务)
type JobLock struct {
	kv         clientv3.KV
	lease      clientv3.Lease
	jobName    string             //任务名
	cancelFunc context.CancelFunc //用于终止自动续租
	leaseId    clientv3.LeaseID
	isLocked   bool
}

func InitJobLock(jobName string, kv clientv3.KV, lease clientv3.Lease) (jobLock *JobLock) {
	jobLock = &JobLock{
		kv:      kv,
		lease:   lease,
		jobName: jobName,
	}
	return
}

//尝试上锁(乐观锁)
func (jobLock JobLock) TryLock() (err error) {
	var (
		cancelCtx      context.Context
		cancelFunc     context.CancelFunc
		leaseGrantResp *clientv3.LeaseGrantResponse
		keepResp       *clientv3.LeaseKeepAliveResponse
		keepChan       <-chan *clientv3.LeaseKeepAliveResponse
		leaseId        clientv3.LeaseID
		txn            clientv3.Txn
		lockKey        string
		txnResp        *clientv3.TxnResponse
	)
	//1.创建租约
	if leaseGrantResp, err = jobLock.lease.Grant(context.TODO(), 5); err != nil {
		return
	}
	//context用于取消自动续租
	cancelCtx, cancelFunc = context.WithCancel(context.TODO())
	leaseId = leaseGrantResp.ID
	//2.自动续租
	if keepChan, err = jobLock.lease.KeepAlive(cancelCtx, leaseId); err != nil {
		goto FAIL
	}
	//3.处理续租应答的协程
	go func() {
		select {
		case keepResp = <-keepChan: //自动续租应答
			if keepResp == nil {
				goto End
			}
		}
	End:
	}()
	//4.创建事务txn
	txn = jobLock.kv.Txn(context.TODO())
	//锁路径
	lockKey = common.JObLockDir + jobLock.jobName
	//5.事务抢锁
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet(lockKey))
	//提交事务
	if txnResp, err = txn.Commit(); err != nil {
		goto FAIL
	}
	//6.成功返回，失败释放租约
	if !txnResp.Succeeded { //锁被占用
		err = common.ERR_LOCK_ALREADY_REQUIRED
		goto FAIL
	}

	//抢锁成功
	jobLock.leaseId = leaseId
	if _, err = jobLock.lease.Revoke(context.TODO(), leaseId); err != nil {
		log.Fatal(err)
	}
	return
FAIL:
	cancelFunc()
	if _, err = jobLock.lease.Revoke(context.TODO(), leaseId); err != nil {
		log.Fatal(err)
	}
	return
}

//释放锁
func (jobLock *JobLock) UnkLock() {
	var err error
	if jobLock.isLocked {
		jobLock.cancelFunc() //取消自动续租
		if _, err = jobLock.lease.Revoke(context.TODO(), jobLock.leaseId); err != nil {
			log.Fatal(err)
		} //释放锁
	}
}
