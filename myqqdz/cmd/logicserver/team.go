package main

import (
	"google.golang.org/protobuf/proto"
	"myqqdz/base/glog"
	pb "myqqdz/tools/logicProto"
	"sync"
	"time"
)

/*
* 匹配的逻辑，应该是匹配好几个玩家后就去拿对应的room
* 房间管理器，房间，玩家，新来的人就去创建一个房间然后等待其他玩家加入
* 定时器查询当前等待的房间，如果人数满了，或者超时则请求一个房间来开启游戏
 */
var (
	roomMge *roomManage
	once          = &sync.Once{}
	waiTime int64 = 10
)

type roomManage struct {
	roomM map[uint64]*pb.Room
	rLuck sync.Mutex
}

//获取单例
func getInstanceRoomMge() *roomManage {
	if roomMge == nil {
		once.Do(func() {
			roomMge = &roomManage{
				roomM: make(map[uint64]*pb.Room),
			}
		})
	}
	return roomMge
}

// 加入等待房间
func (rMge *roomManage) waitRoom(name string, id uint64) bool {

	// 调用rpc交给rCenter匹配
	userList := make([]*pb.User, 0)
	userList = append(userList, &pb.User{
		Id:   id,
		Name: name,
	})
	userL := &pb.UserList{
		Num:      1,
		UserList: userList,
	}

	room, err := getRoom(client, userL)
	if err != nil {
		glog.Error("[getRoom] match ", err)
		return false
	}
	rMge.roomM[room.Id] = room
	return true
}

// 超时判定和转态转发
func (rMge *roomManage) timeActor() {

	curTime := time.Now().Unix()
	for id, val := range rMge.roomM {
		if len(val.UserList) > 20 {
			// 直接开启房间
			startRoom(val)
			delete(rMge.roomM, id)
		} else if curTime-val.StartTime > waiTime {
			// 超时开启房间
			startRoom(val)
			delete(rMge.roomM, id)
		} else {
			// 更新当前房间等待的人数
			broadCast(val)
		}
	}
}

// 广播给匹配中的玩家，当前匹配信息
func broadCast(r *pb.Room) {

	// 需要广播的玩家列表 r.UserList
	res := &pb.Response{
		Status:   3,
		Msg:      "当前匹配中玩家列表",
		UserList: r.UserList,
	}

	out, err := proto.Marshal(res)
	if err != nil {
		glog.Error("Failed to encode address:", err)
		return
	}

	for _, user := range r.UserList {
		cMge.userCli[user.Id].conn.Write(out)
	}
}

// 开启房间
func startRoom(r *pb.Room) {
	res := &pb.Response{
		Status: 0,
		Msg:    "匹配成功",
		Room:   r,
	}

	out, _ := proto.Marshal(res)
	glog.Error("开启房间返回 room: ", res)
	for _, user := range r.UserList {
		cMge.userCli[user.Id].conn.Write(out)
	}
}
