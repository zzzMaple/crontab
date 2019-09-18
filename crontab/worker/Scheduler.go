package worker

import (
	"crontab/crontab/common"
)

//调度者
type Scheduler struct {
	jobEventChan chan *common.JobEvent //etcd任务事件队列
	jobPlanTable map[string]*common.JobSchedulePlan
}

//单例
var (
	G_scheduler *Scheduler
)

//处理任务事件
func (scheduler *Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
	var (
		jobSchedulePlan *common.JobSchedulePlan
		jobExisted      bool
		err             error
	)
	for {
		switch jobEvent.EventType {
		case common.JoBEventSave:
			if jobSchedulePlan, err = common.BuildJobSchedulePlan(jobEvent.Job); err != nil {
				return
			}
			scheduler.jobPlanTable[jobEvent.Job.Name] = jobSchedulePlan
		case common.JobEventDelete:
			if jobSchedulePlan, jobExisted = scheduler.jobPlanTable[jobEvent.Job.Name]; jobExisted {
				delete(scheduler.jobPlanTable, jobEvent.Job.Name)
			}
		}
	}
}

//调度协程
func (scheduler *Scheduler) schedulerLoop() {
	//定时任务 common.Job
	var (
		jobEvent *common.JobEvent
	)
	for {
		//监听任务变化事件 (select-case-chan)
		select {
		case jobEvent = <-scheduler.jobEventChan:
			//对内存中维护的任务列表进行CRUD
			scheduler.handleJobEvent(jobEvent)
		}
	}
}

//推送任务事件
func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	scheduler.jobEventChan <- jobEvent
}

//初始化调度器
func InitScheduler() {
	G_scheduler = &Scheduler{
		jobEventChan: make(chan *common.JobEvent, 1000),
		jobPlanTable: make(map[string]*common.JobSchedulePlan),
	}
	go G_scheduler.schedulerLoop() //启动调度协程
	return
}
