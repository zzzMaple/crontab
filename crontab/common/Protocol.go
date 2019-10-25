package common

import (
	"context"
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

//任务日志
type JobLog struct {
	JobName      string `json:"jobName" bson:"jobName"`
	Err          string `json:"err" bson:"err"`
	Command      string `json:"command" bson:"command"`
	Output       string `json:"output" bson:"output"`             //脚本输出
	PlanTime     int64  `json:"planTime" bson:"planTime"`         //计划开始时间
	ScheduleTime int64  `json:"scheduleTime" bson:"scheduleTime"` //实际调度时间
	StartTime    int64  `json:"startTime" bson:"startTime"`       //任务执行开始时间
	EndTime      int64  `json:"endTime" bson:"endTime"`           //任务执行结束时间
}

//日志批次
type LogBatch struct {
	Logs []interface{} //多条日志
}

//任务执行状态
type JobExecuteInfo struct {
	Job        *Job               //任务信息
	PlanTime   time.Time          //理论调度时间
	RealTime   time.Time          //实际调度时间
	CancelCtx  context.Context    //任务cammand的context
	CancelFunc context.CancelFunc //取消cammand执行的cancel函数
}

//任务日志过滤条件
type JobLogFilter struct {
	JobName string `bson:"jobName"`
}

//任务日志排序规则
type SortLogByStartTime struct {
	SortOrder int `bson:"startTime"` //starTime=-1
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

//从/cron/jobs/job10中抹掉/cron/kill
func ExtractKillName(killKey string) string {
	return strings.TrimPrefix(killKey, JobKillerDir)
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

//构造执行状态信息
func BuildJobExecuteInfo(jobPlan *JobPlan) (jobExecuteInfo *JobExecuteInfo) {
	jobExecuteInfo = &JobExecuteInfo{
		Job:      jobPlan.Job,
		PlanTime: jobPlan.NextTime,
		RealTime: time.Now(),
	}
	jobExecuteInfo.CancelCtx, jobExecuteInfo.CancelFunc = context.WithCancel(context.TODO())
	return
}

//提取Worker的IP
func ExtractWorkerIP(regKey string) string {
	return strings.TrimPrefix(regKey, JobWorkerDir)
}
