package main

import (
	"google.golang.org/protobuf/proto"
	"myqqdz/base/glog"
	pb "myqqdz/tools/logicProto"
	"time"
)

var (
	task   chan Task
	cliMge = getCliMge()
	rMg    = getInstanceRoomMge()
)

type Task struct {
	op    string
	user  *pb.User
	token string
}

func initTask() {
	task = make(chan Task, 128)
	go taskLoop()
}

// 任务循环
func taskLoop() {
	ticker1 := time.NewTicker(time.Second * 1)
	for {
		select {
		case t, ok := <-task:
			if !ok {
				break
			}
			op := t.op
			user := t.user
			switch op {
			case "login":
				login(user.Name, user.Passwd)
			case "account":
				createAccount(user.Name, user.Passwd, user.RePasswd)
			case "getAll":
				getAll(user.Id)
			case "meta":
				match(user.Name, t.token, user.Id)

			default:
				handleError(user.Name)
			}
		case <-ticker1.C:
			rMg.timeActor()
		}
	}
}

func login(name, passwd string) {
	handleFuncLogin(name, passwd)
	ret, userInfo := handleFuncLogin(name, passwd)

	response := &pb.Response{
		Status:   ret,
		Msg:      retFlag[ret],
		Userinfo: userInfo,
	}

	out, err := proto.Marshal(response)
	if err != nil {
		glog.Error("Failed to encode address:", err)
	}
	if ret != 0 {
		userConn[userInfo.Name].Write(out)
	} else {
		cliMge.userCli[userInfo.Id].conn.Write(out)
	}
}

func createAccount(name, passwd, rePasswd string) {
	ret, userInfo := handleFuncCreateAccount(name, passwd, rePasswd)
	response := &pb.Response{
		Status:   ret,
		Msg:      retFlag[ret],
		Userinfo: userInfo,
	}
	glog.Error("[register]", userInfo)
	out, err := proto.Marshal(response)
	if err != nil {
		glog.Error("Failed to encode address:", err)
	}
	if ret != 0 {
		userConn[userInfo.Name].Write(out)
	} else {
		cliMge.userCli[userInfo.Id].conn.Write(out)
	}
}

func getAll(id uint64) {
	out, err := handleFuncGetUserList()
	if err != nil {
		glog.Error("Failed to encode address:", err)
	}
	cliMge.userCli[id].conn.Write(out)
}

func match(name, token string, id uint64) {
	res := handleFuncMeta(name, token, id)
	out, err := proto.Marshal(res)
	if err != nil {
		glog.Error("Failed to encode address:", err)
	}
	cliMge.userCli[id].conn.Write(out)
}
