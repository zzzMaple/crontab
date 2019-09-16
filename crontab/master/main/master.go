package main

import (
	"crontab/crontab/master"
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
	//master -config ./master.json
	flag.StringVar(&confFile, "config", "./master.json", "master json")
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
	if err = master.InitConfig(confFile); err != nil {
		goto ERR
	}
	//任务管理器(JobMgr)
	if err = master.InitJobMgr(); err != nil {
		goto ERR
	}
	//启动Api Http服务
	if err = master.InitApiServer(); err != nil {
		goto ERR
	}
	for {
		time.Sleep(1 * time.Second)
	}
ERR:
	fmt.Println(err)
	return
}
