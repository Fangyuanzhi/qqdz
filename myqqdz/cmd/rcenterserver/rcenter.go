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
	glog.SetLogFile("./rcenter.log")
	configParse(*configs)
	runtime.GOMAXPROCS(runtime.NumCPU() / 2)
	RedisConn = redis2.Setup(redisAddr) // 连接redis

	getrS()
	// 开启任务队列
	initRcTask()

	// 开启gRPC监听服务
	initGrpc()
}
