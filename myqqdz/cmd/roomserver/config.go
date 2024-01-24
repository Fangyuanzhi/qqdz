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
	ip          string      // 本机ip
	rCenterAddr string      //rpc连接
	GameTime    int64       //戏最大时常

	BlockNUM     int           // 网格数量/每行，每列
	BlockFoodNum int           // 每个网格内的食物数量
	BlockSize    float32 = 10  // 每个网格的长宽
	GameMapSize  float32 = 100 // 地图的长宽的一半即200*200

	EatInterval int64          // 分身合并的时间间隔
	protectTime int64   = 2000 // 保护时间
	ActInterval int64   = 101  // 相同操作的时间间隔避免响应多次
	Ky          float32 = 3    // 速度超参数
	Kx          float32 = 3

	RobotNameList []string // 机器名称列表
	RobotNum      int32    // 机器人最大数量
	robotNameNum  int      // 机器人名称数量
)

func configParse(filePath string) {
	RobotNameList = make([]string, 0)
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	// 解析连接通信地址
	{
		redisAddr, err = jsonparser.GetString(file, "dbConfig", "redis")
		if err != nil {
			panic(err)
		}
		port, err = jsonparser.GetInt(file, "room", "port")
		if err != nil {
			panic(err)
		}
		ip, err = jsonparser.GetString(file, "global", "ip")
		if err != nil {
			panic(err)
		}
		rip, er := jsonparser.GetString(file, "rCenter", "ip")
		if er != nil {
			panic(err)
		}
		rPort, e := jsonparser.GetInt(file, "rCenter", "port")
		if e != nil {
			panic(err)
		}
		rCenterAddr = rip + ":" + strconv.FormatInt(rPort, 10)
	}

	// 解析超参数
	{
		var tmp int64
		tmp, err = jsonparser.GetInt(file, "room", "blockNum")
		if err != nil {
			panic(err)
		}
		BlockNUM = int(tmp)
		tmp, err = jsonparser.GetInt(file, "room", "blockFoodNum")
		if err != nil {
			panic(err)
		}
		BlockFoodNum = int(tmp)
		EatInterval, err = jsonparser.GetInt(file, "room", "eatInterval")
		if err != nil {
			panic(err)
		}
		protectTime, err = jsonparser.GetInt(file, "room", "protectTime")
		if err != nil {
			panic(err)
		}
		ActInterval, err = jsonparser.GetInt(file, "room", "actInterval")
		if err != nil {
			panic(err)
		}
		tmp, err = jsonparser.GetInt(file, "room", "robotNum")
		if err != nil {
			panic(err)
		}
		RobotNum = int32(tmp)

		gameSize, er := jsonparser.GetFloat(file, "room", "robotNum")
		if er != nil {
			panic(err)
		}
		GameMapSize = float32(gameSize)
		gameSize, er = jsonparser.GetFloat(file, "room", "blockSize")
		if er != nil {
			panic(err)
		}
		BlockSize = float32(gameSize)
		gameSize, er = jsonparser.GetFloat(file, "room", "ky")
		if er != nil {
			panic(err)
		}
		Ky = float32(gameSize)
		gameSize, er = jsonparser.GetFloat(file, "room", "kx")
		if er != nil {
			panic(err)
		}
		Kx = float32(gameSize)
	}
	// 获取机器人列表
	cb := func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		RobotNameList = append(RobotNameList, string(value))
	}
	jsonparser.ArrayEach(file, cb, "room", "robotName")
	robotNameNum = len(RobotNameList)

}
