package common

import (
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"strings"
	"time"
)

//定时任务
type Job struct {
	Name     string `json:"name"`     //任务名
	Command  string `json:"command"`  //shell命令
	CronExpr string `json:"cronExpr"` //cron表达式
}

//Http接口应答,Response(errno msg data)
type Response struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

//任务变化事件有两种
type JobEvent struct {
	EventType int //SAVE,DELETE
	Job       *Job
}

//任务调度计划
type JobPlan struct {
	Job      *Job                 //任务信息
	Expr     *cronexpr.Expression //解析好的cronexpr表达式
	NextTime time.Time            //下次调度时间
}

//任务执行状态
type JobExecuteInfo struct {
	Job      *Job      //任务信息
	PlanTime time.Time //理论调度时间
	RealTime time.Time //实际调度时间
}

//任务执行结果
type JobExecuteResult struct {
	JobExecuteInfo *JobExecuteInfo //执行状态
	Output         []byte          //脚本输出
	Err            error           //错误原因
	StartTime      time.Time       //启动时间
	EndTime        time.Time       //结束时间

}

//应答方法
func BuildResponse(errno int, msg string, data interface{}) (resp []byte, err error) {
	//1.定义一个response
	var response Response
	//赋值resp
	response.Errno = errno
	response.Msg = msg
	response.Data = data
	//2.序列化
	if resp, err = json.Marshal(response); err != nil {
		return
	}
	return
}

//反序列化Job
func UnpackJob(value []byte) (res *Job, err error) {
	job := new(Job)
	if err = json.Unmarshal(value, &job); err != nil {
		return
	}
	res = job
	return
}

//从etcd的Key中提取任务
//从/cron/jobs/job10中抹掉/cron/jobs
func ExtractJobName(jobKey string) string {
	return strings.TrimPrefix(jobKey, JobSaveDir)
}

//创建任务变化事件
func BuildJobEvent(eventType int, job *Job) (jobEvent *JobEvent) {
	return &JobEvent{
		EventType: eventType,
		Job:       job,
	}
}

//构造任务执行计划
func BuildJobSchedulePlan(job *Job) (jobPlan *JobPlan, err error) {
	var (
		expr *cronexpr.Expression
	)
	if expr, err = cronexpr.Parse(job.CronExpr); err != nil {

	}
	//生成任务调度计划对象
	jobPlan = &JobPlan{
		Job:      job,
		Expr:     expr,
		NextTime: expr.Next(time.Now()),
	}
	return
}

//苟傲执行状态信息
func BuildJobExecuteInfo(jobPlan *JobPlan) (jobExecuteInfo *JobExecuteInfo) {
	jobExecuteInfo = &JobExecuteInfo{
		Job:      jobPlan.Job,
		PlanTime: jobPlan.NextTime,
		RealTime: time.Now(),
	}
	return
}
