package main

import (
	"math/rand"
	"myqqdz/base/glog"
	"myqqdz/base/stl"
	pb "myqqdz/tools/logicProto"
	"net"
	"sync"
	"time"
)

/*
* 一个房间的实例
 */

type Room struct {
	roomId     uint64              `roomId:"房间唯一ID"`
	rMap       *roomMap            `rMap:"房间的地图元素"`
	playConn   map[uint64]net.Conn `plays:"玩家通信列表"`
	gameStatus bool                // 房间运行标识 true 在运行 false未运行
	rLuck      sync.Mutex          `rLuck:"房间锁"`
}

//InitRoom 初始化房间
func (r *Room) InitRoom(user *pb.User, room *pb.Room, ret *pb.InitResponse) {
	// 推迟初始化
	if r.rMap == nil {
		r.initMap()
	}
	uid := user.Id
	ret.Player = r.rMap.CreatePlay(user)
	ret.Player.Uid = uid
	glog.Error("[player] ",ret.Player.Uid,ret.Player.Id,ret.Player.Name)
	r.rMap.players[uid] = InitPlayer(user, ret.Player)
	if len(r.rMap.rRobotList) == 0 {
		glog.Info("InitRoom...   初始化粮食", len(r.rMap.players))
		r.rMap.Id = 1
		// food数量
		r.rMap.gLuck.Lock()
		var foodNum = BlockFoodNum
		for i := 0; i < r.rMap.BlockNum; i++ {
			for j := 0; j < r.rMap.BlockNum; j++ {
				for k := 0; k < foodNum; k++ {
					ty := uint32(rand.Intn(30))
					if ty == 29 {
						ty = 4
					} else {
						ty = uint32(rand.Intn(3)) + 1
					}
					r.rMap.blocks[i][j].foods = append(r.rMap.blocks[i][j].foods,
						r.rMap.NewBall(r.rMap.blocks[i][j].x, r.rMap.blocks[i][j].y, ty, user.SkinId, user.Color))
				}
			}
		}
		r.rMap.foodNum = BlockFoodNum * BlockNUM * BlockNUM

		// 创建机器人
		CreateRobot(RobotNum-room.WaitNum, r.rMap, user)
		r.gameStatus = true
		r.rMap.gLuck.Unlock()
	}

	ret.Quark = r.rMap.blocks[int(ret.Player.X/float32(r.rMap.blockSize))][int(ret.Player.Y/float32(r.rMap.blockSize))].foods
	ret.Player.Nx = ret.Player.X + 0.1
	ret.Player.Ny = ret.Player.Y + 0.1
	ret.Status = 0
	ret.Size = int32(r.rMap.mapSize)
	ret.Msg = "Successful"

	// 更新room信息给rCenter
	notify(room)
}

/*
* 将地图划分成多个小格比如按照地图比例进行划分默认 16*9 那么我就可以划分成 16*9个格子=144个
* wNum,hNum 划分的格子数，globalW，地图长宽的一半先上取整
 */

func (r *Room) initMap() {
	// 划分网格
	r.rMap = &roomMap{
		blocks:         make(map[int]map[int]*BlockMap, 0),
		players:        make(map[uint64]*Player, 0),
		mapSize:        GameMapSize,
		BlockNum:       BlockNUM,
		blockSize:      int(GameMapSize*100/float32(BlockNUM)) * 2,
		startTime:      time.Now().Unix(),
		scoreTop:       &stl.Top{},
		Id:             1,
		robotId:        10,
		rRobotList:     make([]uint64, 0),
		cTask:          make(chan Task, 200),
		foods:          make(map[int]map[int][]*pb.Ball, 100),
		settlementFlag: false,
		foodNum:        0,
	}
	r.rMap.scoreTop.Init()
	for i := 0; i < BlockNUM; i++ {
		r.rMap.blocks[i] = make(map[int]*BlockMap, 0)
		xPos := BlockSize*float32(i) + BlockSize/2
		for j := 0; j < BlockNUM; j++ {
			r.rMap.blocks[i][j] = &BlockMap{
				x:       xPos,
				y:       BlockSize*float32(j) + BlockSize/2,
				size:    BlockSize,
				foods:   make([]*pb.Ball, 0),
				spores:  make(map[uint32]*pb.Ball, 0),
				players: make(map[uint32]*pb.Ball, 0),
			}
			if r.rMap.blocks[i][j] == nil {
				glog.Error("初始化网格 ", i, j, "错误")
			} else if r.rMap.blocks[i][j].players == nil {
				glog.Error("初始化网格 ", i, j, "players错误")
			}
		}
	}
	r.playConn = make(map[uint64]net.Conn)
	r.gameStatus = false
}

// 通知rCenter更新room信息,填充信息到room
func notify(r *pb.Room) {
	rid := r.Id
	rMge := getRoomManage()
	room := rMge.rooms[rid]
	r.PlayNum = int32(len(room.playConn))
	r.WaitNum = int32(len(r.UserList)) - r.PlayNum
	r.UserList = make([]*pb.User, 0)
	for _, play := range room.rMap.players {
		r.UserList = append(r.UserList, play.user)
	}

	roomList := &pb.RoomList{}
	roomList.RoomList = append(roomList.RoomList, r)

	res, err := NotifyRoom(client, roomList)
	if err != nil {
		glog.Error("[Notify]", err)
	}
	glog.Error(res)
}
