package common

const (
	//任务保存目录
	JobSaveDir = "/cron/jobs/"

	//任务强杀目录
	JobKillerDir = "/cron/killer/"

	//保存任务事件
	JoBEventSave = 1
	//删除任务事件
	JobEventDelete = 2
	//强杀任务事件
	JobEventKill = 3
	//任务锁目录
	JobLockDir = "/cron/lock/"
	//节点注册目录
	JobWorkerDir = "/cron/workers/"
)
