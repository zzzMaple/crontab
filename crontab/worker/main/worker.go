package main

import (
	"crontab/crontab/worker"
	"flag"
	"fmt"
	"runtime"
	"time"
)

var (
	confFile string //配置文件路径
)

//解析命令行参数
func initArgs() {
	//worker -config ./worker.json
	flag.StringVar(&confFile, "config", "./worker.json", "worker json")
	flag.Parse()
}

//初始化环境
func initEnv() {
	//设置同时可进行的最大CPU数
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var (
		err error
	)
	//初始化命令行参数
	initArgs()
	//初始化线程
	initEnv()
	//加载配置
	if err = worker.InitConfig(confFile); err != nil {
		goto ERR
	}
	//启动执行器
	if err = worker.InitExecutor(); err != nil {
		goto ERR
	}
	//启动调度器
	if err = worker.InitScheduler(); err != nil {
		goto ERR
	}

	//初始化任务管理器
	if err = worker.InitJobMgr(); err != nil {
		goto ERR
	}
	for {
		time.Sleep(1 * time.Second)
	}
	//
ERR:
	fmt.Println(err.Error())
	return
}
