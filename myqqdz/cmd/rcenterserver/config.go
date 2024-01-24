package main

import (
	"github.com/buger/jsonparser"
	"github.com/gomodule/redigo/redis"
	"io/ioutil"
)

var (
	RedisConn *redis.Pool
	redisAddr string // redis连接地址
	port      int64  // 监听端口
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
	port, err = jsonparser.GetInt(file, "rCenter", "port")
	if err != nil {
		panic(err)
	}

}
