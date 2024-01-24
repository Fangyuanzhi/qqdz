package main

import (
	"flag"
	"fmt"
	"math/rand"
	"myqqdz/base/glog"
	redis2 "myqqdz/base/redis"
	"net"
	"runtime"
	"time"
)

var (
	configs = flag.String("cfg", "D:\\codeProgram\\qqdz\\myqqdz\\config\\config.json", "配置文件")
)

func main() {
	// 设置日志
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Lookup("logtostderr").Value.Set("false")
	glog.SetLogFile("./roomServer.log")

	rand.Seed(time.Now().Unix()) //设置随机初始化种子
	configParse(*configs)        // 解析配置文件

	RedisConn = redis2.Setup(redisAddr)

	// 开启与rCenter的rpc通信
	go initGRpc()

	initTask() // 开启任务管理器

	// 开启监听
	listen()
}

func listen() {
	// 开启监听
	port := ":7001"
	listen, err := net.Listen("tcp", port)
	if err != nil {
		glog.Error("listen error:", err)
		return
	}
	defer glog.Flush()

	defer listen.Close()

	fmt.Println("*****************---------服务开启---------:7001**************")
	for {
		conn, er := listen.Accept()
		if er != nil {
			glog.Error("accept failed, err:", er)
			continue
		}
		go handle(conn) // 每一个连接有一个专门的进程负责
	}

}
