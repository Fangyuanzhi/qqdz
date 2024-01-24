package main

import (
	"bufio"
	"github.com/gomodule/redigo/redis"
	"google.golang.org/protobuf/proto"
	"math"
	"myqqdz/base/glog"
	pb "myqqdz/tools/logicProto"
	"net"
	"strconv"
	"time"
)

var (
	lenMessage uint64 = 0
	mesCnt     uint64 = 0
	ql         uint64 = 0
	qc         uint64 = 0
)

func handle(conn net.Conn) {
	defer conn.Close() // 关闭连接
	rMge := getRoomManage()

	for {
		reader := bufio.NewReader(conn)

		var buf [4096]byte
		n, err := reader.Read(buf[:]) // 读取数据
		if err != nil {
			glog.Error("read from client failed, err:", err)
			break
		}

		// 解析Protobuf
		operator := &pb.Operator{}
		if err = proto.Unmarshal(buf[:n], operator); err != nil {
			glog.Error("解析错误")
			continue
		} else {
			glog.Error(operator.Act, operator.User)
		}

		op := operator.Operator
		user := operator.User
		uid := user.Id
		room := operator.Room
		rid := room.Id
		// 判断房间是否存在
		_, ok := rMge.rooms[rid]
		if !ok {
			if op != "init" {
				glog.Error("[room] 房间ID不存在", rid)
				continue
			}
			rMge.rMLuck.Lock()
			rMge.rooms[rid] = &Room{}
			rMge.rooms[rid].initMap()
			rMge.rMLuck.Unlock()
		}
		r := rMge.rooms[rid]
		// 该房间是否已结算
		if r.gameStatus && time.Now().Unix()-r.rMap.startTime > GameTime {
			if !r.rMap.settlementFlag && op != "init" {
				task := Task{
					operator: "settlement",
					user:     user,
					rid:      rid,
				}
				GlobalTask <- task
				continue
			}
		}

		token := operator.Token
		val, ok2 := r.playConn[uid]
		if !ok2 || val != conn {
			r.playConn[uid] = conn
		}

		//glog.Error("-------------start-------------")
		task := Task{
			operator:  op,
			user:      user,
			rid:       rid,
			userToken: token,
			room:      room,
		}
		if operator.Act != nil {
			task.act = operator.Act
		}
		GlobalTask <- task
	}
}

func handleUpdate(user *pb.User, r *Room, token string, room *pb.Room, act *pb.Act) {
	ret := &pb.UpdateResponse{
		Status: 1,
		Top:    &pb.Top{},
	}
	uid := user.Id
	conn := r.playConn[uid]
	var out []byte
	_, err := RedisConn.Dial()
	//c, err := RedisConn.Dial()
	if err != nil {
		glog.Error(err)
		ret.Msg = err.Error()
		out, err = proto.Marshal(ret)
		conn.Write(out)
		return
	}
	rMap := r.rMap

	//glog.Info("handleUpdate...")
	updateRoom(user, rMap, ret, act)

	// 通过视野返回交互过的网格
	// 获取摄像机的视野
	xL, xR, yB, yU, xNL, xNR, yNB, yNU, cid := r.rMap.PlayView(uid)
	// 获取对应视野的网格
	xLeft, xRight := Bound(int(xL/BlockSize)-3), Bound(int(xR/BlockSize)+3)
	yBottom, yUp := Bound(int(yB/BlockSize)-3), Bound(int(yU/BlockSize)+3)

	//glog.Info("返回对应的网格", xLeft, xRight, yBottom, yUp)
	// 收集要放回的对象
	for i := xLeft; i <= xRight; i++ {
		for j := yBottom; j <= yUp; j++ {
			if rMap.blocks[i][j].players != nil {
				for _, val := range rMap.blocks[i][j].players {
					ret.Players = append(ret.Players, val)
				}
			}
			if rMap.blocks[i][j].spores != nil && len(rMap.blocks[i][j].spores) > 0 {
				for _, val := range rMap.blocks[i][j].spores {
					ret.Spore = append(ret.Spore, val)
				}
			}
			ret.Foods = append(ret.Foods, rMap.blocks[i][j].foods...)
		}
	}
	// 计算视野
	ret.CamPos = rMap.CameraPos(uid, xL, xR, yB, yU, xNL, xNR, yNB, yNU)
	ret.CamPos.Id = cid
	ret.Top.MyTop = rMap.updateTop(uid, user.Name)

	for idx, val := range rMap.scoreTop.PairTop {
		ret.Top.Pair = append(ret.Top.Pair, &pb.Pair{Key: val.Key, Name: val.Name, Val: val.Val})
		if idx >= 9 {
			break
		}
	}
	//glog.Error("handleUpdate successful")
	//glog.Error("返回玩家数: ", len(ret.Players), "孢子数:", len(ret.Spore), "粮食数量: ", len(ret.Foods))
	//glog.Error("返回的视野: ", ret.CamPos.X, ret.CamPos.Y, ret.CamPos.Nx, ret.CamPos.Ny)

	out, err = proto.Marshal(ret)
	if err != nil {
		glog.Error("Failed to encode address:", err)
	}
	//glog.Error("消息长度:", len(out))

	lenMessage += uint64(len(out))
	mesCnt++

	if mesCnt%100 == 0 {
		glog.Error("[conn 长度1]:", lenMessage/mesCnt)
		ql += lenMessage / mesCnt
		qc++
		lenMessage = 0
		mesCnt = 0
		if qc%10 == 0 {
			glog.Error("[conn 长度2]:", ql/qc)
		}

	}

	conn.Write(out)

	// 更新操作后的玩家位置和操作计时
	handlePost(uid, rMap, act)
}

// 更新单个player，与最大质量的交互
func updateRoom(user *pb.User, rMap *roomMap, ret *pb.UpdateResponse, act *pb.Act) {
	uid := user.Id
	tempPlay := rMap.players[uid].play
	if len(tempPlay) == 0 {
		ret.Msg = "玩家位置信息为空！"
		glog.Error("控制球为空,重新生成")
		var templayer = rMap.CreatePlay(user)
		rMap.players[uid] = InitPlayer(user, templayer)
		return
	}

	// 循环判断主球
	var scoreMax float32 = 0
	var idMax uint32
	for idx, val := range tempPlay {
		if val.Score > scoreMax {
			scoreMax = val.Score
			idMax = idx
		}
	}
	// 进行处理
	ballMaster := tempPlay[idMax]
	updateMaster(uid, rMap, ballMaster, act)
	for idx, val := range tempPlay {
		if idx == idMax {
			continue
		}
		updateOther(uid, rMap, ballMaster, val, act)
	}
	// 计算总得分
	rMap.players[uid].score = 0
	for _, val := range rMap.players[uid].play {
		rMap.players[uid].score += val.Score
	}
	ret.Status = 0
	ret.Msg = "Successful"
}

func updateMaster(uid uint64, rMap *roomMap, master *pb.Ball, act *pb.Act) {
	x, y, sc := master.X, master.Y, Score2Size(master.Score)
	glog.Info("处理前玩家坐标：", x, y, master.Nx, master.Ny)
	xPos, yPos := XyFromBlock(x, y)
	blockMaster := rMap.blocks[xPos][yPos]
	// 计算玩家可见的范围内的地图格子,返回的是一个矩形 xL->xR,yB->yU
	xLeft, xRight, yBottom, yUp := ViewGetBlock(x, y, sc)
	// 更新孢子
	for i := xLeft; i <= xRight; i++ {
		for j := yBottom; j <= yUp; j++ {
			if rMap.blocks[i][j].spores != nil && len(rMap.blocks[i][j].spores) > 0 {
				for _, val := range rMap.blocks[i][j].spores {
					if time.Now().UnixMilli()-val.CreateTime < 500 {
						continue
					}
					xs, ys := XyFromBlock(val.X, val.Y)
					xNPos, yNPos := XyFromBlock(val.Nx, val.Ny)
					if xs != xNPos || ys != yNPos {
						rMap.blocks[xNPos][yNPos].spores[val.Id] = val
						delete(rMap.blocks[xs][ys].spores, val.Id)
					}
					val.X = val.Nx
					val.Y = val.Ny
				}
			}
		}
	}
	// 循环每个格子做交互判断
	flagBeEat := false // 玩家被吃不进行下面的操作
	//glog.Info("updateMaster 当前需要判断交互的网格", xLeft, xRight, yBottom, yUp)
	for i := xLeft; i <= xRight; i++ {
		if flagBeEat {
			break
		}
		for j := yBottom; j <= yUp; j++ {
			block := rMap.blocks[i][j]
			// ----------------吞噬游戏玩法逻辑-----------------------
			if rMap.PlayEat(block, blockMaster, master, act) {
				flagBeEat = true
				break
			}
		}
	}
	if flagBeEat {
		return
	}
	// --------------进行动作-----------------
	rMap.DealAct(uid, blockMaster, master, act)
}

func updateOther(uid uint64, rMap *roomMap, master, other *pb.Ball, act *pb.Act) {
	dua := time.Now().UnixMilli() - other.CreateTime
	dua /= 100
	// 模拟向心力 修改运动方向
	mX, mY := master.X, master.Y
	oX, oY := other.X, other.Y
	// 中心距
	dist := float32(math.Sqrt(float64((mX-oX)*(mX-oX) + (mY-oY)*(mY-oY))))
	if dist == 0 {
		dist = 0.01
	}
	// 质量引力,产生的加速度 F=G（M1×M2）/(R^2)
	power := 5 / dist
	if power > 0.7 {
		power = 0.7
	}
	if power < 0.5 {
		power = 0.5
	}
	// 引力方向
	aX, aY := (mX-oX)/dist, (mY-oY)/dist
	other.MoveS = act.MoveS
	radius := Score2Size(master.Score)

	// 修改速度方向
	if dua <= 8 { // 短时间内是互斥的
		other.MoveX = act.MoveX - 3*aX*power
		other.MoveY = act.MoveY - 3*aY*power
	} else if dua > 8 && dua <= 50 { // 中间微弱吸力
		other.MoveX = act.MoveX + 0.5*aX*power
		other.MoveY = act.MoveY + 0.5*aY*power
	} else if dua > 50 && dua < 150 { // 后期微弱的吸引力
		radius *= 1.7
		if dist > radius { // 太远给吸力
			other.MoveX = act.MoveX + 0.1*aX*power
			other.MoveY = act.MoveY + 0.1*aY*power
		} else { // 太近给斥力
			other.MoveX = act.MoveX - 0.1*aX*power
			other.MoveY = act.MoveY - 0.1*aY*power
		}
	} else {
		other.MoveX = act.MoveX + 0.1*aX*power
		other.MoveY = act.MoveY + 0.1*aY*power
	}
	if master.Uid > 1000 {
		glog.Error("[时间] ", dua)
		glog.Error("[速度] ", "中心距：", dist, " aX:", aX, " aY:", aY, " 速度：", act.MoveS)
		glog.Error("[方向] ", act.MoveX, "->", other.MoveX, " ", act.MoveY, "->", other.MoveY)
	}

	updateMaster(uid, rMap, other, &pb.Act{
		MoveX: other.MoveX,
		MoveY: other.MoveY,
		MoveS: other.MoveS,
		Type:  act.Type,
	})
}

func handleInit(user *pb.User, r *Room, token string, room *pb.Room) {
	ret := &pb.InitResponse{
		Status: 1,
	}
	uid := user.Id
	name := user.Name
	conn := r.playConn[uid]
	var out []byte
	_, err := RedisConn.Dial()
	//c, err := RedisConn.Dial()
	if err != nil {
		glog.Error(err)
		ret.Msg = err.Error()
		out, err = proto.Marshal(ret)
		conn.Write(out)
		return
	}

	r.InitRoom(user, room, ret)
	ret.Player.CreateTime = r.rMap.startTime * 1000
	out, err = proto.Marshal(ret)
	if err != nil {
		glog.Error("[protobuf] ", err)
	}
	glog.Info("消息长度：", len(out))
	r.rMap.updateTop(uid, name)
	conn.Write(out)
}

// 结算
func handleSettlement(r *Room) {
	rMap := r.rMap
	rMap.settlementFlag = true
	c, _ := RedisConn.Dial()
	ret := &pb.UpdateResponse{
		Status: -1,
	}
	glog.Error("-----开始结算--------")
	for _, player := range rMap.players {
		uid := player.user.Id
		if uid < 1000 {
			continue
		}
		name := player.user.Name
		score := int(player.score)
		conn := r.playConn[uid]
		sc, e := redis.Ints(c.Do("HMGET", uid, "level", "experience"))
		if e != nil {
			glog.Error(e)
			ret.Status = 1
			ret.Msg = e.Error()
		} else {
			sc[1] += score
			sc[0] = sc[1] / 200
			_, e = c.Do("HMSET", uid, "level", sc[0], "experience", sc[1])
			if e != nil {
				glog.Error(e)
				ret.Status = 1
				ret.Msg = e.Error()
			}
		}
		ret.Msg = "本局分数：" + strconv.Itoa(score)
		ret.User = &pb.UserInfo{
			Name:       name,
			Id:         uid,
			Level:      int32(sc[0]),
			Experience: int32(sc[1]),
		}
		out, err := proto.Marshal(ret)
		if err != nil {
			glog.Error("Failed to encode address:", err)
		}
		conn.Write(out)
	}

	if ret.Status == -1 {
		delete(getRoomManage().rooms, r.roomId)
		r.gameStatus = false
	}
	rMap.settlementFlag = false
}

func handlePost(uid uint64, rMap *roomMap, act *pb.Act) {
	// 更新后可能所在的网格会变化
	for _, val := range rMap.players[uid].play {
		xPos, yPos := XyFromBlock(val.X, val.Y)
		xNPos, yNPos := XyFromBlock(val.Nx, val.Ny)
		if xPos != xNPos || yPos != yNPos {
			rMap.blocks[xNPos][yNPos].players[val.Id] = val
			delete(rMap.blocks[xPos][yPos].players, val.Id)
		}
		val.X = val.Nx
		val.Y = val.Ny
	}
	// 更新动作类型
	rMap.players[uid].preAct = act.Type
	rMap.players[uid].preTime = time.Now().UnixMilli()
}
