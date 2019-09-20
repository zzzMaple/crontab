package worker

import (
	"context"
	"crontab/crontab/common"
	"os/exec"
	"time"
)

//任务执行器
type Executor struct {
}

//单例
var (
	G_executor *Executor
)

//执行一个任务
func (executor Executor) ExecuteJob(info *common.JobExecuteInfo) {
	var (
		cmd     *exec.Cmd
		err     error
		output  []byte
		result  *common.JobExecuteResult
		jobLock *JobLock
	)
	go func() {
		//任务结果
		result = &common.JobExecuteResult{
			JobExecuteInfo: info,
			Output:         make([]byte, 0),
		}
		//初始化分布式锁
		jobLock = G_jobMgr.CreteJobLock(info.Job.Name)
		//记录开始时间
		result.StartTime = time.Now()
		//上锁
		err = jobLock.TryLock()
		//释放锁
		defer jobLock.UnkLock()
		if err != nil { //上锁失败
			result.Err = err
			result.EndTime = time.Now()
		} else {
			//上锁成功后，重置任务启动时间
			result.StartTime = time.Now()
			//执行shell命令
			cmd = exec.CommandContext(context.TODO(), "/bin/bash", "-c", info.Job.Command)
			//执行并捕获输出
			output, err = cmd.CombinedOutput()
			//最终目标是在任务执行后，把执行的结果返回给Scheduler，Scheduler会从executingtable中删除掉执行记录
			result.EndTime = time.Now()
			result.Err = err
			result.Output = output
			//任务结果
		}
		G_scheduler.PushJobReturn(result)

	}()
}

//初始化执行器

func InitExecutor() (err error) {
	G_executor = &Executor{}
	return
}
