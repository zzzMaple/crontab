package worker

import (
	"crontab/crontab/common"
	"fmt"
	"time"
)

//调度者
type Scheduler struct {
	jobEventChan      chan *common.JobEvent             //etcd任务事件队列
	jobPlanTable      map[string]*common.JobPlan        //任务调度表
	jobExecutingTable map[string]*common.JobExecuteInfo //任务执行表
	jobResultChan     chan *common.JobExecuteResult     //任务结果chan
}

//单例
var (
	G_scheduler *Scheduler
)

//处理任务事件
func (scheduler *Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
	var (
		jobPlan    *common.JobPlan
		jobExisted bool
		err        error
	)
	switch jobEvent.EventType {
	case common.JoBEventSave:
		if jobPlan, err = common.BuildJobSchedulePlan(jobEvent.Job); err != nil {
			return
		}
		scheduler.jobPlanTable[jobEvent.Job.Name] = jobPlan
	case common.JobEventDelete:
		if jobPlan, jobExisted = scheduler.jobPlanTable[jobEvent.Job.Name]; jobExisted {
			delete(scheduler.jobPlanTable, jobEvent.Job.Name)
		}
	}
}

//处理任务结果
func (scheduler *Scheduler) handleJobResult(result *common.JobExecuteResult) {
	//删除执行状态
	delete(scheduler.jobExecutingTable, result.JobExecuteInfo.Job.Name)
	fmt.Println("任务执行完成")
}

//尝试执行任务(调度和执行是分开的)
func (scheduler *Scheduler) TryStartJob(jobPlan *common.JobPlan) {
	// 调度 和 执行 是2件事情
	var (
		jobExecuteInfo *common.JobExecuteInfo
		jobExecuting   bool
	)
	// 执行的任务可能运行很久, 1分钟会调度60次，但是只能执行1次, 防止并发！
	// 如果任务正在执行，跳过本次调度
	if jobExecuteInfo, jobExecuting = scheduler.jobExecutingTable[jobPlan.Job.Name]; jobExecuting {
		fmt.Println("尚未退出,跳过执行:", jobPlan.Job.Name)
		return
	}
	// 构建执行状态信息
	jobExecuteInfo = common.BuildJobExecuteInfo(jobPlan)
	// 保存执行状态
	scheduler.jobExecutingTable[jobPlan.Job.Name] = jobExecuteInfo
	// 执行任务
	//fmt.Println("执行任务:", jobExecuteInfo.Job.Name, jobExecuteInfo.PlanTime, jobExecuteInfo.RealTime)
	G_executor.ExecuteJob(jobExecuteInfo)
}

//重新计算任务调度状态
func (scheduler *Scheduler) TrySchedule() (schedulerAfter time.Duration) {
	var (
		jobPlan  *common.JobPlan
		now      time.Time
		nearTime *time.Time
	)
	//当前时间
	now = time.Now()
	//如果任务表为空，则一直睡眠
	if len(scheduler.jobPlanTable) == 0 {
		schedulerAfter = 1 * time.Second
		return
	}
	//1.遍历所有任务
	for _, jobPlan = range scheduler.jobPlanTable {
		//2.过期的任务立即执行
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
			scheduler.TryStartJob(jobPlan)
			fmt.Println("调度任务", jobPlan.Job.Name)
			jobPlan.NextTime = jobPlan.Expr.Next(now) //update nextTime
		}
		//3.统计最近任务的过期时间(N秒后过期==scheduleAfter)
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}
	schedulerAfter = (*nearTime).Sub(now)
	return
}

//调度协程
func (scheduler *Scheduler) schedulerLoop() {
	//定时任务 common.Job
	var (
		jobEvent      *common.JobEvent
		scheduleAfter time.Duration
		scheduleTimer *time.Timer
		jobResult     *common.JobExecuteResult
	)
	//初始化一秒
	scheduleAfter = scheduler.TrySchedule()
	//调度的延迟定时器
	scheduleTimer = time.NewTimer(scheduleAfter)
	for {
		//监听任务变化事件 (select-case-chan)
		select { //select每次只会执行一个case(公平选择)，如若没有case满足执行条件且没有default case, select将会阻塞直到有case满足执行条件
		case jobEvent = <-scheduler.jobEventChan:
			//对内存中维护的任务列表进行CRUD
			scheduler.handleJobEvent(jobEvent)
		case <-scheduleTimer.C: //最近的任务到期了
		case jobResult = <-scheduler.jobResultChan:
			scheduler.handleJobResult(jobResult)
		}
		//调度一次任务
		scheduleAfter = scheduler.TrySchedule()
		//重置调度间隔
		scheduleTimer.Reset(scheduleAfter)
	}
}

//推送任务事件
func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	scheduler.jobEventChan <- jobEvent
}

//初始化调度器
func InitScheduler() (err error) {
	G_scheduler = &Scheduler{
		jobEventChan:      make(chan *common.JobEvent, 1000),
		jobPlanTable:      make(map[string]*common.JobPlan),
		jobExecutingTable: make(map[string]*common.JobExecuteInfo),
		jobResultChan:     make(chan *common.JobExecuteResult, 1000),
	}
	go G_scheduler.schedulerLoop() //启动调度协程
	return
}

//回传任务执行结果
func (scheduler Scheduler) PushJobReturn(jobResult *common.JobExecuteResult) {
	scheduler.jobResultChan <- jobResult
}
