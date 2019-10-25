package master

import (
	"crontab/crontab/common"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"time"
)

//任务Http接口
type ApiServer struct {
	httpServer http.Server
}

//配置单例
var (
	G_apiserver *ApiServer
)

//保存任务接口
func handleJobSave(w http.ResponseWriter, r *http.Request) {
	var (
		err     error
		postJob string
		job     common.Job
		oldJob  *common.Job
		bytes   []byte
	)
	//把任务保存到etcd中
	//1.解析POST表单(Http默认不解析，需主动调用)
	if err = r.ParseForm(); err != nil {
		goto ERR
	}
	//2.取表单中的Job字段
	if postJob = r.Form.Get("job"); postJob == "" {
		goto ERR
	}
	//3.反序列化Job
	if err = json.Unmarshal([]byte(postJob), &job); err != nil {
		goto ERR
	}
	//4.保存到etcd中
	if oldJob, err = G_jobMgr.JobSave(&job); err != nil {
		goto ERR
	}
	//5.返回正常应答{{"error": 0, "msg": "success", "data": {...}}}
	if bytes, err = common.BuildResponse(0, "success", oldJob); err == nil {
		if _, err = w.Write(bytes); err != nil {
			goto ERR
		}
	}
	return
ERR:
	//6.返回异常应答
	if bytes, err = common.BuildResponse(-1, "err", nil); err == nil {
		if _, err = w.Write(bytes); err != nil {
			goto ERR
		}
	}
}

//删除任务接口
func handleJobDelete(w http.ResponseWriter, r *http.Request) {
	var (
		err     error
		jobName string
		oldJob  *common.Job
		bytes   []byte
	)
	//1.解析
	if err = r.ParseForm(); err != nil {
		goto ERR
	}
	//2.取得删除的任务名
	jobName = r.Form.Get("name")
	//3.调用JobMgr的JobDelete方法
	if oldJob, err = G_jobMgr.JobDelete(jobName); err != nil {
		goto ERR
	}
	//4.正常应答(BuildResponse)
	if bytes, err = common.BuildResponse(0, "success", oldJob); err == nil {
		if _, err = w.Write(bytes); err != nil {
			goto ERR
		}
	}
	return
	//错误应答
ERR:
	if bytes, err = common.BuildResponse(-1, "err", nil); err == nil {
		if _, err = w.Write(bytes); err != nil {
			goto ERR
		}
	}
}

//获取任务列表接口
func handleJobList(w http.ResponseWriter, r *http.Request) {
	var (
		err     error
		jobList []*common.Job
		bytes   []byte
	)
	//jobDir = common.JobSaveDir
	//获取任务
	if jobList, err = G_jobMgr.JobList(); err != nil {
		goto ERR
	}
	//正常应答
	if bytes, err = common.BuildResponse(0, "success", jobList); err == nil {
		if _, err = w.Write(bytes); err != nil {
			goto ERR
		}
	}
	return
ERR:
	//错误应答
	if bytes, err = common.BuildResponse(-1, "err", nil); err == nil {
		if _, err = w.Write(bytes); err != nil {
			goto ERR
		}
	}
}

//强制杀死某个任务
func handleJobKill(w http.ResponseWriter, r *http.Request) {
	var (
		err   error
		name  string
		bytes []byte
	)
	//解析POST表单
	if err = r.ParseForm(); err != nil {
		goto ERR
	}
	//设置要kill的任务名
	name = r.Form.Get("name")
	//kill任务
	if err = G_jobMgr.JobKill(name); err != nil {
		goto ERR
	}
	//正常应答
	if bytes, err = common.BuildResponse(0, "success", name); err == nil {
		if _, err = w.Write(bytes); err != nil {
			goto ERR
		}
	}
	return
	//错误应答
ERR:
	if bytes, err = common.BuildResponse(-1, "err", nil); err == nil {
		if _, err = w.Write(bytes); err != nil {
			goto ERR
		}
	}

}

//查询日志
func handleJobLog(w http.ResponseWriter, r *http.Request) {
	var (
		err       error
		name      string //日志名称
		skipParam string
		limitPara string
		skip      int //从第N条开始
		limit     int //返回日志数目
		logArr    []*common.JobLog
		bytes     []byte
	)
	if err = r.ParseForm(); err != nil {
		goto ERR
	}
	name = r.Form.Get("name")
	skipParam = r.Form.Get("skip")
	limitPara = r.Form.Get("limit")
	if skip, err = strconv.Atoi(skipParam); err != nil {
		skip = 0
	}
	if limit, err = strconv.Atoi(limitPara); err != nil {
		limit = 20
	}
	if logArr, err = G_logMgr.ListLog(name, skip, limit); err != nil {
		goto ERR
	}
	//正常应答
	if bytes, err = common.BuildResponse(0, "success", logArr); err == nil {
		if _, err = w.Write(bytes); err != nil {
			goto ERR
		}
	}
	return
	//错误应答
ERR:
	if bytes, err = common.BuildResponse(-1, "err", nil); err == nil {
		if _, err = w.Write(bytes); err != nil {
			goto ERR
		}
	}
}

//获取健康节点IP
func handleWorkerList(w http.ResponseWriter, r *http.Request) {
	var (
		workerIP []string
		err      error
		bytes    []byte
	)

	if workerIP, err = G_workerMgr.workerList(); err != nil {
		goto ERR
	}

	//正常应答
	if bytes, err = common.BuildResponse(0, "success", workerIP); err == nil {
		if _, err = w.Write(bytes); err != nil {
			goto ERR
		}
	}
	return
ERR:
	//错误应答
	if bytes, err = common.BuildResponse(-1, "err", nil); err == nil {
		if _, err = w.Write(bytes); err != nil {
			goto ERR
		}
	}
}

//初始化服务
func InitApiServer() (err error) {
	var (
		mux           *http.ServeMux
		listener      net.Listener
		httpServer    *http.Server
		staticDir     http.Dir //静态文件目录
		staticHandler http.Handler
	)
	//配置路由
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)
	mux.HandleFunc("/job/del", handleJobDelete)
	mux.HandleFunc("/job/list", handleJobList)
	mux.HandleFunc("/job/kill", handleJobKill)
	mux.HandleFunc("/job/log", handleJobLog)
	mux.HandleFunc("/worker/list", handleWorkerList)
	//启动TCP监听
	if listener, err = net.Listen("tcp", ":"+strconv.Itoa(G_config.ApiPort)); err != nil {
		return
	}
	// /index.html

	//静态文件目录
	staticDir = http.Dir(G_config.WebRoot)
	staticHandler = http.FileServer(staticDir)
	mux.Handle("/", http.StripPrefix("/", staticHandler))
	//创建Http服务
	httpServer = &http.Server{
		ReadTimeout:  time.Duration(G_config.ApiReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(G_config.ApiWriteTimeout) * time.Millisecond,
		Handler:      mux,
	}
	//单例模式
	G_apiserver = &ApiServer{
		httpServer: *httpServer,
	}
	//启动服务端
	go httpServer.Serve(listener)

	return
}
