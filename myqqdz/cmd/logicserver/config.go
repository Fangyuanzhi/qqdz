package main

import (
	"github.com/buger/jsonparser"
	"github.com/gomodule/redigo/redis"
	"io/ioutil"
	"strconv"
)

var (
	RedisConn   *redis.Pool // redis连接池
	redisAddr   string      // redis连接地址
	port        int64       // 监听端口
	rCenterAddr string      //rpc连接
)

func configParse(filePath string) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	redisAddr, err = jsonparser.GetString(file, "dbConfig", "redis")
	if err != nil {
		panic(err)
	}
	port, err = jsonparser.GetInt(file, "logic", "port")
	if err != nil {
		panic(err)
	}
	ip, er := jsonparser.GetString(file, "rCenter", "ip")
	if er != nil {
		panic(err)
	}
	rPort, e := jsonparser.GetInt(file, "rCenter", "port")
	if e != nil {
		panic(err)
	}
	rCenterAddr = ip + ":" + strconv.FormatInt(rPort, 10)
}
