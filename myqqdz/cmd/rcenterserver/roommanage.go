package main

import (
	"github.com/gomodule/redigo/redis"
	"math"
	"myqqdz/base/glog"
	pb "myqqdz/tools/logicProto"
	"sync"
	"time"
)

var (
	rS       *RoomServer
	oncerS         = &sync.Once{}
	roomCap  int32 = 20  //	房间最大人数
	roomTime int64 = 120 //房间最长等待匹配时间
)

func getrS() *RoomServer {
	if rS == nil {
		oncerS.Do(func() {
			rS = &RoomServer{
				waitRoom: make(map[uint64]*pb.Room),
				runRoom:  make(map[string]map[uint64]*pb.Room),
			}
		})
	}
	return rS
}

type RoomServer struct {
	waitRoom map[uint64]*pb.Room            `waitRoom:"等待中的房间"`
	runRoom  map[string]map[uint64]*pb.Room `runRoom:"map[roomServer的通信地址][开启的房间唯一id]房间状态"`
	rSLuck   sync.Mutex
}

// 选择房间，先看有没有等待中的房间，有则放入等待中的房间，没有则新创建一个房间
func selectRoom(userList *pb.UserList) (*pb.Room, bool) {
	// 如果已经在房间就需要合并房间
	now := time.Now().Unix()
	room := &pb.Room{}
	if isRoom(room.Id) && now-room.StartTime < roomTime {
		return rS.waitRoom[room.Id], true
	}

	userNum := int32(len(userList.UserList))
	// 有合适的房间直接放入房间
	for rid, r := range rS.waitRoom {
		if now-r.StartTime > roomTime*2 {
			rS.runRoom[r.Addr][rid] = rS.waitRoom[rid]
			delete(rS.waitRoom, rid)
			continue
		}
		if now-r.StartTime > roomTime {
			continue
		}
		if r.WaitNum+userNum < roomCap {
			r.WaitNum += userNum
			r.UserList = append(r.UserList, userList.UserList...)
			return r, true
		}
	}

	// 没有合适的房间,创建新等待房间
	c := RedisConn.Get()
	rid, err := redis.Uint64(c.Do("INCR", "roomId"))
	if err != nil {
		glog.Error("[redis] ", err)
		return room, false
	}

	addr, fl := selectRoomServer()
	if !fl {
		glog.Error("[roomServer] 没有找到合适的roomServer,请开启roomServer！")
		return room, false
	}
	rS.waitRoom[rid] = &pb.Room{
		Id:        rid,
		WaitNum:   userNum,
		Addr:      addr,
		UserList:  userList.UserList,
		StartTime: time.Now().Unix(),
	}
	return rS.waitRoom[rid], true
}

// 更新当前的运行的房间和等待房间
func updaterS(rList *pb.RoomList) {
	for _, room := range rList.RoomList {
		if room.Id == 0 {
			_, ok := rS.runRoom[room.Addr]
			if !ok {
				rS.runRoom[room.Addr] = make(map[uint64]*pb.Room)
				glog.Error("[register] ", room.Addr, " 已经注册")
			}
			continue
		}
		rid := room.Id
		_, ok := rS.waitRoom[rid]
		if ok {
			if rS.runRoom[room.Addr] == nil {
				rS.runRoom[room.Addr] = make(map[uint64]*pb.Room)
			}
			rS.waitRoom[rid].StartTime = room.StartTime
			rS.waitRoom[rid].PlayNum = room.PlayNum
			rS.waitRoom[rid].WaitNum = 0
			rS.runRoom[room.Addr][rid] = rS.waitRoom[rid]
			delete(rS.waitRoom, rid)
			continue
		}
		rS.runRoom[room.Addr][rid] = room
	}
}

// 从当前roomServer中选择负载最低的
func selectRoomServer() (string, bool) {
	maxRs := math.MaxInt
	var ipPort string
	for addr, roomS := range rS.runRoom {
		if len(roomS) < maxRs {
			ipPort = addr
			maxRs = len(roomS)
		}
	}
	if maxRs == math.MaxInt {
		return ipPort, false
	}

	return ipPort, true
}

// 判断房间是否存在
func isRoom(rid uint64) bool {
	_, ok := rS.waitRoom[rid]
	if ok {
		return true
	}
	return false
}
