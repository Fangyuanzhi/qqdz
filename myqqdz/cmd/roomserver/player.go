package main

import (
	pb "myqqdz/tools/logicProto"
	"sync"
)

type Player struct {
	user    *pb.User
	score   float32
	play    map[uint32]*pb.Ball
	preAct  int32 // 上一次收到的动作类型
	preTime int64 // 上一次动作的时间
	pLock   sync.Mutex
}

//InitPlayer 创建一个Player对象
func InitPlayer(user *pb.User, play *pb.Ball) *Player {
	pl := &Player{
		user:  user,
		score: play.Score,
		play:  make(map[uint32]*pb.Ball, 0),
	}
	pl.play[play.Id] = play
	return pl
}
