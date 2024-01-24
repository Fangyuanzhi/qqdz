package main

import (
	"bufio"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"google.golang.org/protobuf/proto"
	"myqqdz/base/glog"
	redis2 "myqqdz/base/redis"
	pb "myqqdz/tools/logicProto"
	"net"
	"strconv"
)

var (
	NamePassError  int32 = 10101
	RedisConnError int32 = 10102
	RedisCommError int32 = 10103
	NotFoundError  int32 = 10104
	FoundError     int32 = 10105
	retFlag        map[int32]string
	userConn       map[string]net.Conn // 预登陆名单
)

func initDefine() {
	retFlag = make(map[int32]string, 8)
	retFlag[NamePassError] = "输入有误"
	retFlag[RedisConnError] = "数据库连接失败"
	retFlag[RedisCommError] = "数据库命令有误"
	retFlag[NotFoundError] = "用户名或密码错误"
	retFlag[FoundError] = "用户已经存在"
	retFlag[0] = "登录成功！"
	userConn = make(map[string]net.Conn)
}

func initTcp() {
	initDefine()
	listen, err := net.Listen("tcp", ":"+strconv.FormatInt(port, 10))
	if err != nil {
		glog.Error("listen error:", err)
		return
	}

	fmt.Println("---------服务开启---------")
	for {
		conn, err := listen.Accept()
		if err != nil {
			glog.Error("accept failed, err:", err)
			continue
		}
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			glog.Error("[coon] ", err)
		}
	}(conn) // 关闭连接
	for {
		reader := bufio.NewReader(conn)
		var buf [1024]byte
		n, err := reader.Read(buf[:]) // 读取数据
		if err != nil {
			glog.Error("read from client failed, err:", err)
			break
		}
		operator := &pb.Operator{}
		if err = proto.Unmarshal(buf[:n], operator); err != nil {
			glog.Error("解析错误")
		} else {
			glog.Info("[conn]", conn.RemoteAddr().String(), operator)
		}

		user := operator.User
		val, ok := userConn[user.Name]
		if !ok || val != conn {
			userConn[user.Name] = conn
		}
		op := operator.Operator
		playTask := Task{
			op:    op,
			user:  user,
			token: operator.Token,
		}
		task <- playTask
	}
}

// 登录
func handleFuncLogin(name, passwd string) (int32, *pb.UserInfo) {
	user := &pb.UserInfo{
		Name:  name,
		Token: "",
	}
	if !check(name, passwd, "") {
		glog.Error("用户名或密码格式错误")
		return NamePassError, user
	}

	c, err := RedisConn.Dial()
	if err != nil {
		glog.Error(err)
		return RedisConnError, user
	}

	{
		s1, e := redis.Int(c.Do("HEXISTS", name, "passwd"))
		if e != nil {
			glog.Error(e)
			return RedisCommError, user
		}
		if s1 < 1 {
			glog.Error("用户名不存在：", name)
			return NotFoundError, user
		}
		s2, er := redis.String(c.Do("HGET", name, "passwd"))
		if er != nil {
			glog.Error(er)
			return RedisCommError, user
		}
		if s2 != passwd {
			glog.Error("用户密码错误：", name)
			return NamePassError, user
		}
		// 设置状态
		_, er = redis.Int(c.Do("HSET", name, "status", 1))
		if er != nil {
			glog.Error(er)
			return RedisCommError, user
		}

		// 设置token
		token := redis2.GetToken(16)
		_, er = redis.Int(c.Do("HSET", name, "token", token))
		if er != nil {
			glog.Error(er)
			return RedisCommError, user
		}
		user.Token = token
		id, err := redis.Uint64(c.Do("HGET", name, "id"))
		if err != nil {
			glog.Error("[redis] ", e)
			return RedisCommError, user
		}
		// 获取等级，经验
		sc, e := redis.Ints(c.Do("HMGET", id, "level", "experience"))
		if e != nil {
			glog.Error("[redis] ", e)
			return RedisCommError, user
		}

		user.Id = id
		user.Level = int32(sc[0])
		user.Experience = int32(sc[1])
	}
	cliMge := getCliMge()
	cliMge.addUser(user.Name, user.Token, userConn[user.Name], user.Id)
	return 0, user
}

// 创建账号处理函数
func handleFuncCreateAccount(name, passwd1, passwd2 string) (int32, *pb.UserInfo) {
	user := &pb.UserInfo{
		Name:       name,
		Token:      "",
		Level:      1,
		Experience: 0,
	}
	if !check(name, passwd1, passwd2) {
		glog.Error("用户名或密码格式错误")
		return NamePassError, user
	}

	c, err := RedisConn.Dial()
	if err != nil {
		glog.Error(err)
		return RedisConnError, user
	}
	var id uint64
	{
		s1, e := redis.Int(c.Do("HEXISTS", name, "passwd"))
		if e != nil {
			glog.Error(e)
			return RedisCommError, user
		}
		if s1 > 0 {
			glog.Error("用户已经存在：", name)
			return FoundError, user
		}
		id, err = redis.Uint64(c.Do("INCR", "userid"))
		if err != nil {
			glog.Error(err)
			return RedisCommError, user
		}
		user.Id = id
		token := redis2.GetToken(16)
		_, er := redis.String(c.Do("HMSET", name, "passwd", passwd1, "status", 1, "token", token, "id", id))
		_, er = redis.String(c.Do("HMSET", id, "name", name, "level", 1, "experience", 0))
		if er != nil {
			glog.Error(er)
			return RedisCommError, user
		}
		user.Token = token
	}
	_, er := redis.Int(c.Do("SADD", "playerList", id))
	if er != nil {
		glog.Error(er)
		return RedisCommError, user
	}
	cliMge := getCliMge()
	cliMge.addUser(user.Name, user.Token, userConn[user.Name], user.Id)
	return 0, user
}

// 获取当前所有玩家列表 处理函数
func handleFuncGetUserList() ([]byte, error) {
	response := &pb.Response{
		Status:   1,
		Msg:      "获取玩家列表失败",
		UserList: []*pb.User{},
	}
	c, err := RedisConn.Dial()
	if err != nil {
		glog.Error(err)
		response.Msg = err.Error()
		return proto.Marshal(response)
	}

	s2, er := redis.Strings(c.Do("SMEMBERS", "playerList"))
	if er != nil {
		glog.Error(er)
		response.Msg = er.Error()
		return proto.Marshal(response)
	}
	fmt.Println(s2)
	for _, str := range s2 {
		response.UserList = append(response.UserList, &pb.User{
			Name: str,
		})
	}
	out, e := proto.Marshal(response)
	return out, e
}

// 匹配 这个是同步的逻辑，改成广播的逻辑就没什么了
func handleFuncMeta(name, token string, id uint64) *pb.Response {
	// 调用team进行匹配，这个只回传告诉客户端正在匹配
	// 验证token
	response := &pb.Response{
		Status: 2,
		Msg:    "successful",
	}

	//c, err := RedisConn.Dial()
	//if err != nil {
	//	glog.Error(err)
	//	response.Msg = err.Error()
	//	return response
	//}
	// 验证token
	//tk, e := redis.String(c.Do("HGET", name, "token"))
	//if e != nil {
	//	glog.Error("[redis] ", e)
	//	response.Msg = e.Error()
	//	return response
	//}
	//if tk != token {
	//	glog.Error("[token]", "token认证失败")
	//	response.Msg = err.Error()
	//	return response
	//}

	if rMg.waitRoom(name, id) {
		response.Status = 1
	} else {
		response.Msg = "匹配失败，未知错误"
	}

	return response
}

func handleError(name string) {
	response := &pb.Response{
		Status: 2,
		Msg:    "操作符错误",
	}
	out, err := proto.Marshal(response)
	if err != nil {
		glog.Error("Failed to encode address:", err)
	}
	userConn[name].Write(out)
}

// 检查用户名、密码是否合理
func check(name, passwd1, passwd2 string) bool {
	if len(name) > 14 || len(name) < 1 {
		glog.Error("用户名格式错误")
		return false
	}
	if len(passwd1) > 24 || len(passwd1) < 6 {
		glog.Error("用户密码格式错误")
		return false
	}
	if len(passwd2) > 0 && passwd1 != passwd2 {
		return false
	}
	return true
}
