package main

import (
	"math/rand"
	"myqqdz/base/glog"
	pb "myqqdz/tools/logicProto"
	"time"
)

var (
	GlobalTask chan Task
)

func initTask() {
	if GlobalTask == nil {
		GlobalTask = make(chan Task, 128)
	}
	go taskLoop()
}

type Task struct {
	operator  string
	user      *pb.User
	rid       uint64
	userToken string
	room      *pb.Room
	act       *pb.Act
}

func taskLoop() {
	rMge := getRoomManage()
	tickerRobot := time.NewTicker(time.Microsecond * 1200)
	tickerFood := time.NewTicker(time.Microsecond * 800)
	for {
		select {
		case task, ok := <-GlobalTask:
			if !ok {
				break
			}
			op := task.operator
			uid := task.user.Id
			rid := task.rid
			r := rMge.rooms[rid]

			if r == nil {
				glog.Error("r 为空")
				continue
			}
			if r.rMap == nil {
				glog.Error("地图 为空")
				break
			}
			if r.rMap.settlementFlag {
				glog.Info("[room] ", rid, "在结算")
				break
			}

			switch op {
			case "init":
				if time.Now().Unix()-r.rMap.startTime > GameTime {
					conn, _ := r.playConn[uid]
					r.initMap()
					r.playConn[uid] = conn
					glog.Error("重新初始化")
				}
				handleInit(task.user, r, task.userToken, task.room)
			case "update":
				handleUpdate(task.user, r, task.userToken, task.room, task.act)
			case "settlement":
				handleSettlement(r)
			}

		case <-tickerRobot.C:
			// 操控机器人,修改不要一次就所有room都更新
			rand.Seed(time.Now().Unix())
			globalUpdateRobot()
		case <-tickerFood.C:
			// 生成全图粮食
			globalGeneratorFood()
		}
	}
}

func globalGeneratorFood() {
	rMge := getRoomManage()
	for _, room := range rMge.rooms {
		rMap := room.rMap
		if len(rMap.players)-len(rMap.rRobotList) == 0 || rMap.settlementFlag {
			continue
		}
		rMap.GeneratorFood()
	}
}

func globalUpdateRobot() {
	rMge := getRoomManage()
	for _, room := range rMge.rooms {
		if !room.gameStatus {
			continue
		}
		rMap := room.rMap
		if len(rMap.players)-len(rMap.rRobotList) == 0 || rMap.settlementFlag {
			continue
		}
		UpdateRobot(rMap)
	}
}
