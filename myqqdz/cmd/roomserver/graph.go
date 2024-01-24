package main

import (
	"math"
	"math/rand"
	"myqqdz/base/glog"
	"myqqdz/base/stl"
	pb "myqqdz/tools/logicProto"
	"sync"
	"time"
)

/*  全局地图
*	包含W*H的地图，划分成W*H个小格子，每次都只更新玩家可见的地图,
*	地图元素包含
*	food 粮食元素， 不可动，随着生成器刷新，具有随机性 由生成器和触发器管理
*	stab 刺球，可以吃孢子变大，会被推动，可被玩家吃下后炸刺
*	spore 孢子，吐出来会动一会，之后就保持不变
*   play 玩家，会移动和交互，确定视野
 */

type roomMap struct {
	blocks    map[int]map[int]*BlockMap
	players   map[uint64]*Player // 玩家列表
	mapSize   float32            // 地图尺寸 长宽的一半
	blockSize int                // 网格尺寸 float32*100
	BlockNum  int                // 地图划分的网格

	startTime int64                      // 游戏时长
	foods     map[int]map[int][]*pb.Ball // 新增的食物
	foodNum   int                        // 当前食物数量
	scoreTop  *stl.Top                   // 排行榜

	Id         uint32   // 自增ID
	robotId    uint64   //机器人自增id
	rRobotList []uint64 // 机器人列表

	settlementFlag bool //	房间结算标志

	cTask chan Task
	gLuck sync.Mutex
}

/*
*	地图中的一个小方格
*  	是全图的组件
 */

type BlockMap struct {
	x    float32 // 中心坐标
	y    float32
	size float32 // 宽高

	foods   []*pb.Ball          //地图元素粮食,包括刺球
	spores  map[uint32]*pb.Ball //地图元素孢子
	players map[uint32]*pb.Ball //地图元素玩家
}

//GeneratorFood 生成全局food
func (rMap *roomMap) GeneratorFood() {
	maxFoodNum := BlockFoodNum * BlockNUM * BlockNUM
	if len(rMap.players) == 0 {
		return
	}

	cnt := rMap.foodNum
	if cnt < maxFoodNum {
		num := (maxFoodNum - cnt) / rMap.BlockNum / rMap.BlockNum / 20
		cnt = 0
		if num < 1 {
			num = 1
		}
		for i := 0; i < rMap.BlockNum; i++ {
			for j := 0; j < rMap.BlockNum; j++ {
				for k := 0; k < num; k++ {
					ty := uint32(rand.Intn(5000))
					if ty == 4999 {
						ty = 4
					} else {
						ty = uint32(rand.Intn(3)) + 1
					}
					rMap.blocks[i][j].foods = append(rMap.blocks[i][j].foods,
						rMap.NewBall(rMap.blocks[i][j].x, rMap.blocks[i][j].y, ty, 0, &pb.Color{
							R: 1, G: 1, B: 1, A: 255,
						}))
				}
				cnt += len(rMap.blocks[i][j].foods)
			}
		}
		rMap.foodNum = cnt
		glog.Info("当前网格所含粮食数量", cnt)
	}
}

//NewBall 生成新对象中心点,每个格子的长度
func (rMap *roomMap) NewBall(x, y float32, ty uint32, skinId uint32, color *pb.Color) *pb.Ball {
	rMap.Id++
	size := rMap.blockSize
	var score float32 = 1
	switch ty {
	case 0: // 初始化玩家
		score = 50.0
	case 4: //刺球
		score = 80.0
	}
	x = BoundXY(x + float32(rand.Intn(size)-size/2)/100)
	y = BoundXY(y + float32(rand.Intn(size)-size/2)/100)

	if ty < 5 && ty != 0 {
		return &pb.Ball{
			X:     x,
			Y:     y,
			Score: score,
			Type:  ty,
			Id:    rMap.Id,
		}
	}

	return &pb.Ball{
		X:          x,
		Y:          y,
		Nx:         x,
		Ny:         y,
		Score:      score,
		Type:       ty,
		Id:         rMap.Id,
		SkinId:     skinId,
		Color:      color,
		CreateTime: time.Now().UnixMilli(),
	}
}

//CreatePlay 在地图上随机创建一个玩家，并加入地图
func (rMap *roomMap) CreatePlay(user *pb.User) *pb.Ball {
	x, y := float32(rand.Intn(2*int(rMap.mapSize))), float32(rand.Intn(2*int(rMap.mapSize)))
	play := rMap.NewBall(x, y, 0, user.SkinId, user.Color)
	glog.Error("[time] ", play.CreateTime)
	play.Name = user.Name
	play.Uid = user.Id
	xPos, yPos := XyFromBlock(play.X, play.Y)
	rMap.blocks[xPos][yPos].players[play.Id] = play

	return play
}

// 更新并获取排行榜
func (rMap *roomMap) updateTop(uid uint64, name string) int32 {
	return rMap.scoreTop.Update(stl.Pair{
		Key:  uid,
		Name: name,
		Val:  rMap.players[uid].score,
	})
}

/*
* 以下为玩法的逻辑
 */

//PlayEat 吞噬的逻辑
func (rMap *roomMap) PlayEat(block, blockMaster *BlockMap, play *pb.Ball, act *pb.Act) bool {
	uid := play.Uid
	// 吃粮食和刺球
	foods := make([]*pb.Ball, 0)
	for _, val := range block.foods {
		if Comp(play.X, play.Y, play.Score, val.X, val.Y, val.Score, 1.0, 0.9) > 0 {
			rMap.foodNum--
			play.Score += val.Score

			if val.Type == 4 {
				rMap.EatStab(uid, blockMaster, play, act)
			}
		} else {
			foods = append(foods, val)
		}
	}

	block.foods = foods

	// 吃孢子
	for _, val := range block.spores {
		if Comp(play.X, play.Y, play.Score, val.Nx, val.Ny, val.Score, 1.0, 0.9) > 0 {
			play.Score += val.Score
			delete(block.spores, val.Id)
		}
	}

	blockMaster.players[play.Id] = play
	// 玩家之间
	flagBeEat := false
	for _, val := range block.players {
		if val.Id == play.Id { // 玩家被吃不进行下面的操作
			continue
		}

		// 玩家自身的小球
		if val.Uid == uid {
			compare := Comp(play.X, play.Y, play.Score, val.X, val.Y, val.Score, 1, 0.9)
			if compare > 0 && time.Now().UnixMilli()-val.CreateTime > EatInterval {
				play.Score += val.Score
				delete(block.players, val.Id)
				delete(rMap.players[val.Uid].play, val.Id)
			} else if compare < 0 && time.Now().UnixMilli()-play.CreateTime > EatInterval {
				val.Score += play.Score
				delete(blockMaster.players, play.Id)
				delete(rMap.players[uid].play, play.Id)
				flagBeEat = true
				break
			}
			continue
		}
		// 玩家和其他玩家
		_, ok := rMap.players[val.Uid]
		if !ok {
			glog.Error("[players] 不能存在该rMap.players[val.Uid] ", val.Uid)
			continue
		}
		compare := Comp(play.X, play.Y, play.Score, val.X, val.Y, val.Score, 1.1, 0.9)
		if compare > 0 && !(len(rMap.players[val.Uid].play) == 1 && time.Now().UnixMilli()-val.CreateTime <= protectTime) {
			play.Score += val.Score
			delete(block.players, val.Id)
			delete(rMap.players[val.Uid].play, val.Id)
		} else if compare < 0 && !(len(rMap.players[uid].play) == 1 && time.Now().UnixMilli()-play.CreateTime <= protectTime) {
			val.Score += play.Score
			delete(blockMaster.players, play.Id)
			delete(rMap.players[uid].play, play.Id)
			flagBeEat = true
			break
		}
	}
	return flagBeEat
}

//EatStab 吃刺
func (rMap *roomMap) EatStab(uid uint64, blockMaster *BlockMap, play *pb.Ball, act *pb.Act) {
	count := len(rMap.players[uid].play)
	if count >= 8 {
		return
	}
	if count > 2 {
		count = 8 - count
	} else {
		count = 6
	}

	score := play.Score / (float32(count) + 1.5)
	radius := float64(Score2Size(score)) * 1.0 //推算创生坐标和速度
	// 速度修正
	speed := AdoptSpeed(score)
	for i := 0; i < count; i++ {
		rMap.Id++
		sx := math.Sin(math.Pi * float64(i) / float64(count) * 2)
		cy := math.Cos(math.Pi * float64(i) / float64(count) * 2)
		tmpx := math.Min(0.3, float64(act.MoveS*speed)*sx)
		tmpy := math.Min(0.3, float64(act.MoveS*speed)*cy)
		xi := BoundXY(play.X + float32(radius*0.4*sx))
		yi := BoundXY(play.Y + float32(radius*0.4*cy))
		newPlayer := &pb.Ball{
			Id:         rMap.Id,
			X:          xi,
			Y:          yi,
			Score:      score,
			Type:       0,
			MoveX:      act.MoveX,
			MoveY:      act.MoveY,
			MoveS:      act.MoveS * speed * 1.2,
			Nx:         BoundXY(xi + float32(radius*tmpx)),
			Ny:         BoundXY(yi + float32(radius*tmpy)),
			CreateTime: time.Now().UnixMilli(),
			Uid:        uid,
			Name:       play.Name,
			SkinId:     play.SkinId,
			Color:      play.Color,
		}
		// 新增玩家操控的球
		rMap.players[uid].play[newPlayer.Id] = newPlayer
		// 球更新到对应的block中
		xPos, yPos := XyFromBlock(newPlayer.Nx, newPlayer.Ny)
		rMap.blocks[xPos][yPos].players[newPlayer.Id] = newPlayer
	}
	play.Score = score * 1.5
	blockMaster.players[play.Id] = play
	rMap.players[uid].play[play.Id] = play
}

//DealAct 处理玩家操作
func (rMap *roomMap) DealAct(uid uint64, masterBlock *BlockMap, master *pb.Ball, act *pb.Act) {

	if act.Type != 0 {
		tu := time.Now().UnixMilli() - rMap.players[uid].preTime
		if tu < ActInterval {
			act.Type = 0
		}
	}

	tempPlay := rMap.players[uid].play
	x, y := master.X, master.Y
	mx, my := float64(act.MoveX), float64(act.MoveY)
	distXY := float32(math.Sqrt(mx*mx + my*my))
	if distXY == 0 {
		distXY = 0.01
	}
	absX, absY := act.MoveX/distXY, act.MoveY/distXY

	switch act.Type {
	case 1: // 吐袍子
		if master.Score <= 20 {
			break
		}
		master.Score -= 10
		radius := Score2Size(master.Score) //推算创生坐标和速度
		xx := x + radius*absX
		yy := y + radius*absY
		rMap.Id++
		newSpore := &pb.Ball{
			Id:         rMap.Id,
			X:          BoundXY(xx),
			Y:          BoundXY(yy),
			Score:      10,
			Type:       5,
			MoveX:      Float2(act.MoveX),
			MoveY:      Float2(act.MoveY),
			MoveS:      Float2(act.MoveS * 3),
			Nx:         BoundXY(xx + absX*4.5),
			Ny:         BoundXY(yy + absY*4.5),
			SkinId:     master.SkinId,
			Color:      master.Color,
			CreateTime: time.Now().UnixMilli(),
		}
		masterBlock.players[master.Id] = master
		xPos, yPos := XyFromBlock(newSpore.Nx, newSpore.Ny) // 当前球所在的区域
		rMap.blocks[xPos][yPos].spores[newSpore.Id] = newSpore
	case 2: // 分身
		// 达数量上限
		if master.Score <= 40 || len(tempPlay) >= 8 {
			break
		}
		// 新分裂出的分身
		if time.Now().UnixMilli()-master.CreateTime < 101 {
			break
		}
		breakScore := master.Score * 0.46
		master.Score *= 0.54
		masterBlock.players[master.Id] = master
		radius := Score2Size(master.Score) //推算创生坐标和速度
		// 速度修正
		speed := AdoptSpeed(master.Score)
		xx := x + radius*absX
		yy := y + radius*absY

		glog.Error("[分身] ", act, "半径: ", radius, "score: ", master.Score, "方向 [", absX, ", ", absY, "]")
		radius += 5
		rMap.Id++
		newPlayer := &pb.Ball{
			Id:         rMap.Id,
			X:          BoundXY(xx),
			Y:          BoundXY(yy),
			Score:      breakScore,
			Type:       0,
			MoveX:      Float2(act.MoveX),
			MoveY:      Float2(act.MoveY),
			MoveS:      Float2(act.MoveS * speed),
			Nx:         BoundXY(xx + absX*radius*0.8),
			Ny:         BoundXY(yy + absY*radius*0.8),
			CreateTime: time.Now().UnixMilli(),
			Uid:        uid,
			Name:       master.Name,
			SkinId:     master.SkinId,
			Color:      master.Color,
		}
		glog.Error("[分身] ", newPlayer.X, newPlayer.Y, newPlayer.Nx, newPlayer.Ny)
		masterBlock.players[master.Id] = master
		rMap.players[uid].play[newPlayer.Id] = newPlayer
		xPos, yPos := XyFromBlock(newPlayer.Nx, newPlayer.Ny) // 当前球所在的区域
		rMap.blocks[xPos][yPos].players[newPlayer.Id] = newPlayer

	default:
		// 未知操作 todo
	}

	master.MoveX = act.MoveX
	master.MoveY = act.MoveY
	// 速度修正
	speed := AdoptSpeed(master.Score)
	speed = float32(math.Sqrt(float64(speed)))
	//speed = 1
	master.MoveS = act.MoveS * speed
	if act.Type == 0 {
		if master.MoveS > 1.5 {
			master.MoveS = 1.5
		} else if master.MoveS < 0.2 {
			master.MoveS = 0.2
		}
		master.Nx = BoundXY(master.Nx + act.MoveX*master.MoveS/Kx)
		master.Ny = BoundXY(master.Ny + act.MoveY*master.MoveS/Ky)
		return
	}
	master.Nx = BoundXY(x + act.MoveX*master.MoveS/Kx)
	master.Ny = BoundXY(y + act.MoveY*master.MoveS/Ky)
}

//PlayView 玩家视野大小
func (rMap *roomMap) PlayView(uid uint64) (float32, float32,
	float32, float32, float32, float32, float32, float32, uint32) {
	tempPlay := rMap.players[uid].play
	// 各个球体到中心的最大边界上下左右边界
	var boundUp, boundDown, boundLeft, boundRight float32 = 0, 2 * GameMapSize, 2 * GameMapSize, 0
	var boundNUp, boundNDown, boundNLeft, boundNRight float32 = 0, 2 * GameMapSize, 2 * GameMapSize, 0

	rMap.players[uid].score = 0
	var scoreMax float32 = 0
	var id uint32 = 0
	var centerX, centerY float32
	for idx, val := range tempPlay {
		rMap.players[uid].score += val.Score
		if val.Score > scoreMax {
			scoreMax = val.Score
			centerX, centerY = val.X, val.Y
			id = idx
		}
		ScSize := Score2Size(val.Score)
		if val.Y+ScSize > boundUp {
			boundUp = val.Y + ScSize
		}
		if val.Y-ScSize < boundDown {
			boundDown = val.Y - ScSize
		}
		if val.X+ScSize > boundRight {
			boundRight = val.X + ScSize
		}
		if val.X-ScSize < boundLeft {
			boundLeft = val.X - ScSize
		}

		if val.Ny+ScSize > boundNUp {
			boundNUp = val.Y + ScSize
		}
		if val.Ny-ScSize < boundNDown {
			boundNDown = val.Ny - ScSize
		}
		if val.Nx+ScSize > boundNRight {
			boundNRight = val.Nx + ScSize
		}
		if val.Nx-ScSize < boundNLeft {
			boundNLeft = val.Nx - ScSize
		}
	}

	w := float32(math.Max(float64(centerX-boundLeft), float64(boundRight-centerX)))
	h := float32(math.Max(float64(centerY-boundDown), float64(boundUp-centerY)))
	if h*16/9 > w {
		w = h * 16 / 9
	}
	boundLeft, boundRight = centerX-w, centerX+w
	boundDown, boundUp = centerY-h, centerY+h
	return boundLeft, boundRight, boundDown, boundUp, BoundXY(boundNLeft), BoundXY(boundNRight), BoundXY(boundNDown), BoundXY(boundNUp), id
}

//CameraPos 计算相机坐标
func (rMap *roomMap) CameraPos(uid uint64, boundLeft, boundRight, boundDown,
	boundUp, boundNLeft, boundNRight, boundNDown, boundNUp float32) *pb.CameraPos {

	// 边界的中心
	boundX, boundY := float32(boundLeft+boundRight)/2, float32(boundUp+boundDown)/2
	boundNX, boundNY := float32(boundNLeft+boundNRight)/2, float32(boundNUp+boundNDown)/2

	// 类比
	camepos := &pb.CameraPos{
		X:       boundX,
		Y:       boundY,
		Nx:      boundNX,
		Ny:      boundNY,
		Sx:      boundX - boundLeft + BlockSize,
		Sy:      boundY - boundDown + BlockSize,
		Nsx:     boundNX - boundNLeft + BlockSize,
		Nsy:     boundNY - boundNDown + BlockSize,
		Score:   rMap.players[uid].score,
		Curtime: time.Now().Unix(),
	}

	return camepos
}
