package main

import (
	"math"
	"math/rand"
	"myqqdz/base/glog"
	pb "myqqdz/tools/logicProto"
)

/*
*  robot应该和玩家操作比较相同才好写代码，可以使用robot+自增的ID来作为机器人名称
*  更新机器人就需要一个额外的机器人名称列表来作为键值来进行对应的操作
 */

//CreateRobot 创建机器人
func CreateRobot(num int32, rMap *roomMap, user *pb.User) {
	n := int(num)
	temp := make(map[string]uint64)
	temp[user.Name] = user.Id
	for i := 0; i < n; i++ {
		x, y := float32(rand.Intn(2*int(rMap.mapSize))), float32(rand.Intn(2*int(rMap.mapSize)))
		color := &pb.Color{
			R: uint32(rand.Intn(255)), G: uint32(rand.Intn(255)), B: uint32(rand.Intn(255)),
			A: 255,
		}
		play := rMap.NewBall(x, y, 0, 0, color)

		name := RobotNameList[rand.Intn(robotNameNum)]
		if _, ok := temp[name]; ok {
			i--
			continue
		}
		temp[name] = rMap.robotId
		play.Name = name
		play.Uid = rMap.robotId
		user := &pb.User{
			Id:     rMap.robotId,
			Name:   name,
			SkinId: 0,
			Color:  color,
		}
		rMap.robotId++
		rMap.players[play.Uid] = InitPlayer(user, play)
		glog.Error("[robot]", play.Uid, play.Id, play.Name)
		// 放入块中
		xPos, yPos := XyFromBlock(play.X, play.Y)
		rMap.blocks[xPos][yPos].players[play.Id] = play
		rMap.rRobotList = append(rMap.rRobotList, play.Uid)
	}
	glog.Error("[robot] ", n, len(rMap.rRobotList))
}

//RobotOperator 生成机器人操作
func RobotOperator() *pb.Act {
	opr := rand.Intn(300)
	moveFx, moveFy, moveFS := float64(rand.Intn(101)-50), float64(rand.Intn(101)-50), float64(rand.Intn(4))/8+0.05
	if moveFx == 0 {
		moveFx = 10
	}
	if moveFy == 0 {
		moveFy = 10
	}
	moveF := math.Sqrt(moveFx*moveFx + moveFy*moveFy)

	mX := moveFx / moveF
	mY := moveFy / moveF
	act := &pb.Act{
		Type:  0,
		MoveX: float32(mX),
		MoveY: float32(mY),
		MoveS: float32(moveFS),
	}

	if opr < 5 {
		act.Type = 1 // 吐孢子
	} else if opr == 299 {
		act.Type = 2 // 分身
	}
	//glog.Error("[robot act] ", act)
	return act
}

//UpdateRobot 执行机器人操作
func UpdateRobot(rMap *roomMap) {
	ret := &pb.UpdateResponse{
		Status: 1,
	}

	for idx, uid := range rMap.rRobotList {
		if len(rMap.players[uid].play) == 0 {
			rMap.rRobotList = append(rMap.rRobotList[:idx], rMap.rRobotList[idx:]...)
			continue
		}
		if rand.Intn(2) == 0 {
			continue
		}
		user := rMap.players[uid].user
		act := RobotOperator()
		updateRoom(user, rMap, ret, act)
		rMap.updateTop(uid, user.Name)
		// 更新操作后的玩家位置和操作计时
		handlePost(uid, rMap, act)
	}
}
