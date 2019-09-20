package worker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

//程序配置
type Config struct {
	EtcdEndPotints  []string `json:"etcdEndPoints"`
	EtcdDialTimeout int      `json:"etcdDialTimeout"`
}

var (
	//单例
	G_config *Config
)

//初始化配置函数
func InitConfig(filename string) (err error) {
	var (
		content []byte
		config  Config
	)
	//1.读取配置文件
	if content, err = ioutil.ReadFile(filename); err != nil {
		return
	}
	//2.json反序列化
	if err = json.Unmarshal(content, &config); err != nil {
		return
	}
	//3.赋值单例
	G_config = &config
	fmt.Print(G_config)
	return
}
