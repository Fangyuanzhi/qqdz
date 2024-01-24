package main

import "math"

/*
* 一些数学上的判断公式和坐标判断
 */

//AdoptSpeed 修正速度
func AdoptSpeed(score float32) float32 {
	if score <= 40 {
		return float32(math.Sqrt(float64(2.1 * (1 - score/(40+score)))))
	}
	if score <= 160 {
		return float32(math.Sqrt(1.05 - math.Sqrt(float64(score-40)/120)*0.15))
	}
	if score <= 640 {
		return float32(math.Sqrt(0.9 - math.Sqrt(float64(score-160)/480)*0.15))
	}
	return float32(math.Sqrt(float64(0.75 * (1 - score/(30000+score)))))
}

//Comp 比较是否可以吞噬
func Comp(x, y, sc, x1, y1, sc1, k, kd float32) int {
	dist := math.Sqrt(float64((x-x1)*(x-x1) + (y-y1)*(y-y1)))
	diam := math.Sqrt(float64(sc)) * 0.2
	diam1 := math.Sqrt(float64(sc1)) * 0.2
	if sc*k < sc1 && diam1 > diam*float64(kd)+dist {
		return -1
	} else if sc > sc1*k && diam > diam1*float64(kd)+dist {
		return 1
	}
	return 0
}

//Score2Size 计算玩家小球的得分对应的尺寸大小
func Score2Size(score float32) float32 {
	if score > 30000 {
		score = 30000
	}
	return Float2(float32(math.Sqrt(float64(score)*0.042 + 0.15)))
}

//XyFromBlock x,y坐标判断属于哪个网格
func XyFromBlock(x, y float32) (int, int) {
	x, y = BoundXY(x), BoundXY(y)
	return int(x / BlockSize), int(y / BlockSize)
}

// ViewGetBlock x,y,size 判断当前视野所包含的网格 中心的框高
func ViewGetBlock(x, y, size float32) (int, int, int, int) {
	xLeft, xRight := int(BoundXY(x-size)/BlockSize), int(BoundXY(x+size)/BlockSize)
	yBottom, yUp := int(BoundXY(y-size)/BlockSize), int(BoundXY(y+size)/BlockSize)

	return xLeft, xRight, yBottom, yUp
}

//BoundXY 限制坐标不超过边界
func BoundXY(pos float32) float32 {
	if pos >= GameMapSize*2 {
		return GameMapSize*2 - 0.01
	}
	if pos <= 0 {
		return 0.01
	}
	return Float2(pos)
}

func Bound(x int) int {
	if x < 0 {
		return 0
	}
	if x > 19 {
		return 19
	}
	return x
}

func Float2(f float32) float32 {
	a := float32(int(f*100)) / 100
	return a
}
