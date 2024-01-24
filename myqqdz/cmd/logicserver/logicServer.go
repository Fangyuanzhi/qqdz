package main

import (
	"flag"
	"myqqdz/base/glog"
	redis2 "myqqdz/base/redis"
	"runtime"
)

var (
	configs = flag.String("config", "D:\\codeProgram\\qqdz\\myqqdz\\config\\config.json", "json配置文件地址")
)

func main() {
	// 设置日志
	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("false")
	glog.SetLogFile("./logicServer.log")
	configParse(*configs)
	RedisConn = redis2.Setup(redisAddr) // 连接redis
	runtime.GOMAXPROCS(runtime.NumCPU() / 2)
	// 开启任务队列
	initTask()

	// 开启gRPC监听服务
	go initGRpc()

	// 开启socket Tcp监听
	initTcp()
}
