package main

import (
	pb "myqqdz/tools/logicProto"
)

/*
*   匹配玩家-有两个队列：waitRoom和runRoom
*	先在等待队列等待，人数够了后在开启对应的房间
 */

var (
	rcTask *rCenterTasks
)

type rCenterTasks struct {
	rCTasks chan Task
}

type Task struct {
	op   string
	room *pb.Room
}

func initRcTask() {
	rcTask = &rCenterTasks{
		rCTasks: make(chan Task, 128),
	}
	//go TaskLoop()
}

//func TaskLoop() {
//	for {
//		select {
//		case task, ok := <-rcTask.rCTasks:
//			if !ok {
//				break
//			}
//			switch task.op {
//			case "register":
//				registerRoom(task.room)
//			case "update":
//				updateRoom(task.room)
//			case "logoff":
//				delRoom(task.room)
//			}
//			//case <-time.After(time.Microsecond * 390):
//		}
//	}
//}
